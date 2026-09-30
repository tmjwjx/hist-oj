<template>
  <el-card>
    <div slot="header" class="header">
      <span class="panel-title">{{ $t('m.Reg_Admin_Info_Title') }}</span>
      <div class="header-actions">
        <el-button
          size="small"
          type="primary"
          icon="el-icon-download"
          :loading="exporting"
          :disabled="loading || !registrations.length"
          @click="exportExcel"
        >
          {{ exporting ? $t("m.Reg_Exporting") : $t("m.Export_Excel") }}
        </el-button>
        <el-button size="small" @click="$router.back()">{{ $t('m.Back') }}</el-button>
      </div>
    </div>
    <el-table :data="registrations" v-loading="loading" border>
      <el-table-column prop="username" :label="$t('m.Reg_OJ_Username')" min-width="150" fixed></el-table-column>
      <el-table-column prop="uid" :label="$t('m.Reg_User_UID')" min-width="220"></el-table-column>
      <el-table-column
        v-for="field in enabledFields"
        :key="field.value"
        :prop="field.value"
        :label="$t('m.' + field.labelKey)"
        min-width="120"
      ></el-table-column>
      <el-table-column prop="gmtCreate" :label="$t('m.Reg_Time')" min-width="170"></el-table-column>
      <el-table-column :label="$t('m.Operation')" width="90" fixed="right">
        <template slot-scope="scope">
          <el-button type="text" size="small" @click="openEdit(scope.row)">{{ $t('m.Edit') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :title="$t('m.Reg_Edit_Title')" :visible.sync="editVisible" width="520px">
      <el-form v-if="editForm" label-width="90px" size="small">
        <el-form-item :label="$t('m.Reg_OJ_Username')">
          <el-input :value="editForm.username" disabled></el-input>
        </el-form-item>
        <el-form-item
          v-for="field in enabledFields"
          :key="field.value"
          :label="$t('m.' + field.labelKey)"
          required
        >
          <el-select v-if="field.value === 'gender'" v-model="editForm.gender" style="width: 100%">
            <el-option :label="$t('m.Male')" value="男"></el-option>
            <el-option :label="$t('m.Female')" value="女"></el-option>
          </el-select>
          <el-input v-else v-model="editForm[field.value]" :maxlength="field.maxLength"></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="editVisible = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="saveEdit">{{ $t('m.Save') }}</el-button>
      </span>
    </el-dialog>
  </el-card>
</template>

<script>
import api from "@/common/api";
import myMessage from "@/common/message";
import * as XLSX from "xlsx";

// 模块级映射表只存 i18n 键名，渲染时通过 $t('m.' + labelKey) 翻译
const FIELDS = [
  { value: "name", labelKey: "Reg_Field_Name" },
  { value: "class", labelKey: "Reg_Field_Class" },
  { value: "college", labelKey: "Reg_Field_College" },
  { value: "studentId", labelKey: "Reg_Field_Student_ID" },
  { value: "gender", labelKey: "Reg_Field_Gender" },
  { value: "qq", labelKey: "Reg_Field_QQ" },
  { value: "phone", labelKey: "Reg_Field_Phone" },
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
        myMessage.error(this.$t("m.Reg_Load_Failed"));
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
        myMessage.warning(this.$t("m.Reg_Nothing_To_Export"));
        return;
      }
      this.exporting = true;
      try {
        const columns = [
          { key: "username", label: this.$t("m.Reg_OJ_Username") },
          { key: "uid", label: this.$t("m.Reg_User_UID") },
          ...this.enabledFields.map((field) => ({ key: field.value, label: this.$t("m." + field.labelKey) })),
          { key: "status", label: this.$t("m.Reg_Status") },
          { key: "gmtCreate", label: this.$t("m.Reg_Time") },
          { key: "gmtModified", label: this.$t("m.Reg_Last_Modified") },
        ];
        const rows = this.registrations.map((registration) =>
          columns.reduce((row, column) => {
            const value = registration[column.key];
            row[column.label] = column.key === "status"
              ? (value === 1 ? this.$t("m.Reg_Status_Invalid") : this.$t("m.Reg_Status_Normal"))
              : (value == null ? "" : value);
            return row;
          }, {})
        );
        const worksheet = XLSX.utils.json_to_sheet(rows);
        worksheet["!cols"] = columns.map((column) => ({ wch: Math.max(12, column.label.length + 4) }));
        const workbook = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(workbook, worksheet, this.$t("m.Reg_Admin_Info_Title"));
        const cid = this.$route.params.contestId;
        const title = (this.contest && this.contest.title) || this.$t("m.ContestAdm_Contest_Label") + cid;
        const filename = `${title.replace(/[\\/:*?"<>|]/g, "_")}_${this.$t("m.Reg_Admin_Info_Title")}.xlsx`;
        XLSX.writeFile(workbook, filename);
        myMessage.success(this.$t("m.Reg_Export_Success", { count: rows.length }));
      } catch (e) {
        myMessage.error(this.$t("m.Reg_Export_Failed"));
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
          throw new Error(res.data.msg || this.$t("m.Reg_Update_Failed"));
        }
        const index = this.registrations.findIndex((item) => item.id === this.editForm.id);
        if (index >= 0) {
          this.$set(this.registrations, index, { ...this.registrations[index], ...this.editForm });
        }
        this.editVisible = false;
        myMessage.success(this.$t("m.Reg_Update_Success"));
      } catch (e) {
        myMessage.error(e.message || this.$t("m.Reg_Update_Failed"));
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
