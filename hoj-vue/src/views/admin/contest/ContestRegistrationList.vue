<template>
  <el-card>
    <div slot="header" class="header">
      <span class="panel-title">比赛报名信息</span>
      <div class="header-actions">
        <el-button
          size="small"
          type="primary"
          icon="el-icon-download"
          :loading="exporting"
          :disabled="loading || !registrations.length"
          @click="exportExcel"
        >
          {{ exporting ? "正在导出" : "导出 Excel" }}
        </el-button>
        <el-button size="small" @click="$router.back()">返回</el-button>
      </div>
    </div>
    <el-table :data="registrations" v-loading="loading" border>
      <el-table-column prop="username" label="OJ名称" min-width="150" fixed></el-table-column>
      <el-table-column prop="uid" label="用户 UID" min-width="220"></el-table-column>
      <el-table-column
        v-for="field in enabledFields"
        :key="field.value"
        :prop="field.value"
        :label="field.label"
        min-width="120"
      ></el-table-column>
      <el-table-column prop="gmtCreate" label="报名时间" min-width="170"></el-table-column>
      <el-table-column label="操作" width="90" fixed="right">
        <template slot-scope="scope">
          <el-button type="text" size="small" @click="openEdit(scope.row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog title="编辑报名信息" :visible.sync="editVisible" width="520px">
      <el-form v-if="editForm" label-width="90px" size="small">
        <el-form-item label="OJ名称">
          <el-input :value="editForm.username" disabled></el-input>
        </el-form-item>
        <el-form-item
          v-for="field in enabledFields"
          :key="field.value"
          :label="field.label"
          required
        >
          <el-select v-if="field.value === 'gender'" v-model="editForm.gender" style="width: 100%">
            <el-option label="男" value="男"></el-option>
            <el-option label="女" value="女"></el-option>
          </el-select>
          <el-input v-else v-model="editForm[field.value]" :maxlength="field.maxLength"></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveEdit">保存</el-button>
      </span>
    </el-dialog>
  </el-card>
</template>

<script>
import api from "@/common/api";
import myMessage from "@/common/message";
import * as XLSX from "xlsx";

const FIELDS = [
  { value: "name", label: "姓名" },
  { value: "class", label: "班级" },
  { value: "college", label: "学院" },
  { value: "studentId", label: "学号" },
  { value: "gender", label: "性别" },
  { value: "qq", label: "QQ" },
  { value: "phone", label: "电话号码" },
];

export default {
  name: "ContestRegistrationList",
  data() {
    return {
      loading: false,
      exporting: false,
      saving: false,
      editVisible: false,
      editForm: null,
      contest: null,
      registrations: [],
    };
  },
  computed: {
    enabledFields() {
      const configured = this.contest && this.contest.registrationFields;
      let values = Array.isArray(configured) ? configured : [];
      if (typeof configured === "string") {
        try {
          values = JSON.parse(configured || "[]");
        } catch (e) {
          values = [];
        }
      }
      return FIELDS.filter((field) => values.includes(field.value));
    },
  },
  created() {
    this.loadData();
  },
  methods: {
    async loadData() {
      this.loading = true;
      try {
        const cid = this.$route.params.contestId;
        const [contestRes, registrationsRes] = await Promise.all([
          api.admin_getContest(cid),
          api.admin_getContestRegistrations(cid),
        ]);
        this.contest = contestRes.data.data;
        this.registrations = registrationsRes.data.data || [];
      } catch (e) {
        myMessage.error("加载比赛报名信息失败");
      } finally {
        this.loading = false;
      }
    },
    openEdit(row) {
      this.editForm = { ...row };
      this.editVisible = true;
    },
    exportExcel() {
      if (!this.registrations.length) {
        myMessage.warning("暂无报名记录可导出");
        return;
      }
      this.exporting = true;
      try {
        const columns = [
          { key: "username", label: "OJ名称" },
          { key: "uid", label: "用户 UID" },
          ...this.enabledFields.map((field) => ({ key: field.value, label: field.label })),
          { key: "status", label: "报名状态" },
          { key: "gmtCreate", label: "报名时间" },
          { key: "gmtModified", label: "最后修改时间" },
        ];
        const rows = this.registrations.map((registration) =>
          columns.reduce((row, column) => {
            const value = registration[column.key];
            row[column.label] = column.key === "status"
              ? (value === 1 ? "失效" : "正常")
              : (value == null ? "" : value);
            return row;
          }, {})
        );
        const worksheet = XLSX.utils.json_to_sheet(rows);
        worksheet["!cols"] = columns.map((column) => ({ wch: Math.max(12, column.label.length + 4) }));
        const workbook = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(workbook, worksheet, "报名信息");
        const cid = this.$route.params.contestId;
        const title = (this.contest && this.contest.title) || `比赛${cid}`;
        const filename = `${title.replace(/[\\/:*?"<>|]/g, "_")}_报名信息.xlsx`;
        XLSX.writeFile(workbook, filename);
        myMessage.success(`已导出 ${rows.length} 条报名记录`);
      } catch (e) {
        myMessage.error("导出报名信息失败，请稍后重试");
      } finally {
        this.exporting = false;
      }
    },
    async saveEdit() {
      this.saving = true;
      try {
        const res = await api.admin_updateContestRegistration({
          id: this.editForm.id,
          cid: this.editForm.cid,
          name: this.editForm.name,
          class: this.editForm.class,
          college: this.editForm.college,
          studentId: this.editForm.studentId,
          gender: this.editForm.gender,
          qq: this.editForm.qq,
          phone: this.editForm.phone,
        });
        if (res.data && res.data.status !== 200) {
          throw new Error(res.data.msg || "报名信息更新失败");
        }
        const index = this.registrations.findIndex((item) => item.id === this.editForm.id);
        if (index >= 0) {
          this.$set(this.registrations, index, { ...this.registrations[index], ...this.editForm });
        }
        this.editVisible = false;
        myMessage.success("报名信息已更新");
      } catch (e) {
        myMessage.error(e.message || "报名信息更新失败");
      } finally {
        this.saving = false;
      }
    },
  },
};
</script>

<style scoped>
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.el-table {
  width: 100%;
}
</style>
