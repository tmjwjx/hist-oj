<template>
  <div>
    <el-card>
      <div slot="header">
        <span class="panel-title home-title">{{ $t('m.Website_Config') }}</span>
      </div>
      <el-form
        label-position="left"
        label-width="110px"
        ref="form"
        :model="websiteConfig"
      >
        <el-row :gutter="20">
          <el-col :md="8" :xs="24">
            <el-form-item :label="$t('m.Base_Url')" required>
              <el-input
                v-model="websiteConfig.baseUrl"
                :placeholder="$t('m.Base_Url')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="8" :xs="24">
            <el-form-item :label="$t('m.Web_Name')" required>
              <el-input
                v-model="websiteConfig.name"
                :placeholder="$t('m.Web_Name')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="8" :xs="24">
            <el-form-item :label="$t('m.Short_Name')" required>
              <el-input
                v-model="websiteConfig.shortName"
                :placeholder="$t('m.Short_Name')"
              ></el-input>
            </el-form-item>
          </el-col>

          <el-col :md="12" :xs="24">
            <el-form-item :label="$t('m.Record_Name')" required>
              <el-input
                v-model="websiteConfig.recordName"
                :placeholder="$t('m.Record_Name')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="12" :xs="24">
            <el-form-item :label="$t('m.Record_Url')" required>
              <el-input
                v-model="websiteConfig.recordUrl"
                :placeholder="$t('m.Record_Url')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="12" :xs="24">
            <el-form-item :label="$t('m.Project_Name')" required>
              <el-input
                v-model="websiteConfig.projectName"
                :placeholder="$t('m.Project_Name')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="12" :xs="24">
            <el-form-item :label="$t('m.Project_Url')" required>
              <el-input
                v-model="websiteConfig.projectUrl"
                :placeholder="$t('m.Project_Url')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="24" :xs="24">
            <el-form-item :label="$t('m.Web_Desc')" required>
              <el-input
                type="textarea"
                :placeholder="$t('m.Web_Desc')"
                v-model="websiteConfig.description"
                maxlength="150"
                show-word-limit
              >
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :md="24" :xs="24">
            <el-form-item :label="$t('m.Allow_Register')" label-width="120px">
              <el-switch
                v-model="websiteConfig.register"
                active-color="#13ce66"
                inactive-color="#ff4949"
              >
              </el-switch>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <el-button
        type="primary"
        @click.native="saveWebsiteConfig"
        size="small"
        >{{ $t('m.Save') }}</el-button
      >
    </el-card>

    <el-card style="margin-top:15px">
      <div slot="header"><span class="panel-title home-title">AI 验题配置</span></div>
      <el-alert
        title="系统已内置完整验题提示词"
        description="默认覆盖题面、约束、样例推演、标准程序、普通/SPJ/交互模式、全部测试点正式结果与每点 stderr，并统一使用 xhigh 思考水平；此处可继续追加学校自己的验题规则。"
        type="info" :closable="false" show-icon style="margin-bottom:16px"
      />
      <el-form label-width="120px" :model="aiConfig">
        <el-row :gutter="20">
          <el-col :md="12" :xs="24"><el-form-item label="启用 AI"><el-switch v-model="aiConfig.enabled" /></el-form-item></el-col>
          <el-col :md="12" :xs="24"><el-form-item label="模型"><el-input v-model="aiConfig.model" /></el-form-item></el-col>
          <el-col :md="12" :xs="24"><el-form-item label="接口 URL"><el-input v-model="aiConfig.apiUrl" placeholder="可填写 OpenAI 兼容的 /v1 基础地址或完整地址" /></el-form-item></el-col>
          <el-col :md="12" :xs="24"><el-form-item label="API Key"><el-input v-model="aiConfig.apiKey" type="password" show-password /></el-form-item></el-col>
          <el-col :md="8" :xs="24"><el-form-item label="单请求超时"><el-input v-model.number="aiConfig.timeoutSeconds" type="number" :min="1" :max="3600" :step="30"><template slot="append">秒</template></el-input><div style="color:#909399;font-size:12px;line-height:20px">允许 1–3600 秒，保存后立即用于新的 AI 请求</div></el-form-item></el-col>
          <el-col :md="24" :xs="24"><el-form-item label="系统提示词"><el-input v-model="aiConfig.systemPrompt" type="textarea" :rows="8" /></el-form-item></el-col>
          <el-col :md="24" :xs="24"><el-form-item label="验题提示词"><el-input v-model="aiConfig.validationPrompt" type="textarea" :rows="10" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <el-button type="primary" size="small" @click="saveAIConfig">{{ $t('m.Save') }}</el-button>
    </el-card>

    <el-card style="margin-top:15px">
      <div slot="header">
        <span class="panel-title home-title">{{
          $t('m.Home_Rotation_Chart')
        }}</span>
      </div>

      <ul class="el-upload-list el-upload-list--picture-card">
        <li
          tabindex="0"
          class="el-upload-list__item is-ready"
          v-for="(img, index) in carouselImgList"
          :key="index"
        >
          <div>
            <img
              :src="img.url"
              alt="load faild"
              style="height:146px;width:146x"
              class="el-upload-list__item-thumbnail"
            /><span class="el-upload-list__item-actions">
              <span
                class="el-upload-list__item-preview"
                @click="handlePictureCardPreview(img)"
              >
                <i class="el-icon-zoom-in"></i>
              </span>
              <span
                v-if="!disabled"
                class="el-upload-list__item-delete"
                @click="handleDownload(img)"
              >
                <i class="el-icon-download"></i>
              </span>
              <span
                v-if="!disabled"
                class="el-upload-list__item-delete"
                @click="handleRemove(img, index)"
              >
                <i class="el-icon-delete"></i>
              </span>
            </span>
          </div>
        </li>
      </ul>

      <el-upload
        action="/api/file/upload-carouse-img"
        list-type="picture-card"
        accept="image/gif,image/jpeg,image/jpg,image/png,image/svg,image/jfif,image/webp"
        :on-preview="handlePictureCardPreview"
        :on-remove="handleRemove"
        style="display: inline;"
      >
        <i class="el-icon-plus"></i>
      </el-upload>
      <el-dialog :visible.sync="dialogVisible">
        <img width="100%" :src="dialogImageUrl" alt="" />
      </el-dialog>
      <el-dialog :visible.sync="dialogVisible">
        <img width="100%" :src="dialogImageUrl" alt="" />
      </el-dialog>
    </el-card>

    <el-card style="margin-top:15px">
      <div slot="header">
        <span class="panel-title home-title">{{ $t('m.SMTP_Config') }}</span>
      </div>
      <el-form label-position="left" label-width="80px" :model="smtp">
        <el-row :gutter="20">
          <el-col :md="24" :xs="24">
            <el-form-item :label="$t('m.Host')" required label-width="80px">
              <el-input
                v-model="smtp.emailHost"
                :placeholder="$t('m.Host')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="24" :xs="24">
            <el-form-item :label="$t('m.Port')" required label-width="80px">
              <el-input
                v-model="smtp.emailPort"
                type="number"
                :placeholder="$t('m.Port')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="12" :xs="24">
            <el-form-item :label="$t('m.Email')" required label-width="80px">
              <el-input
                v-model="smtp.emailUsername"
                :placeholder="$t('m.Email')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="12" :xs="24">
            <el-form-item :label="$t('m.Password')" label-width="80px" required>
              <el-input
                v-model="smtp.emailPassword"
                type="password"
                :placeholder="$t('m.Password')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="24" :xs="24">
            <el-form-item :label="$t('m.Email_BG')" label-width="80px" required>
              <el-input
                v-model="smtp.emailBGImg"
                :placeholder="$t('m.Email_BG_Desc')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="24" :xs="24">
            <el-form-item label="SSL">
              <el-switch v-model="smtp.emailSsl"> </el-switch>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <el-button type="primary" @click.native="saveSMTPConfig" size="small">{{
        $t('m.Save')
      }}</el-button>
      <el-button
        type="warning"
        @click.native="testSMTPConfig"
        v-if="saved"
        :loading="loadingBtnTest"
        size="small"
        >{{ $t('m.Send_Test_Email') }}</el-button
      >
    </el-card>

    <el-card style="margin-top:15px">
      <div slot="header">
        <span class="panel-title home-title">{{
          $t('m.DataSource_Config')
        }}</span>
      </div>
      <el-form label-position="top" :model="databaseConfig">
        <el-row :gutter="20">
          <el-col :md="12" :xs="24">
            <el-form-item
              :label="'MySQL ' + $t('m.Host')"
              required
              label-width="80px"
            >
              <el-input
                v-model="databaseConfig.dbHost"
                :placeholder="'MySQL ' + $t('m.Host')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="12" :xs="24">
            <el-form-item
              :label="'MySQL ' + $t('m.Port')"
              required
              label-width="80px"
            >
              <el-input
                type="number"
                v-model="databaseConfig.dbPort"
                :placeholder="'MySQL ' + $t('m.Port')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="12" :xs="24">
            <el-form-item
              :label="'MySQL ' + $t('m.Username')"
              required
              label-width="80px"
            >
              <el-input
                v-model="databaseConfig.dbUsername"
                :placeholder="'MySQL ' + $t('m.Username')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="12" :xs="24">
            <el-form-item
              :label="'MySQL ' + $t('m.Password')"
              label-width="80px"
              required
            >
              <el-input
                v-model="databaseConfig.dbPassword"
                type="password"
                :placeholder="'MySQL ' + $t('m.Password')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="8" :xs="24">
            <el-form-item
              :label="'Redis ' + $t('m.Host')"
              label-width="80px"
              required
            >
              <el-input
                v-model="databaseConfig.redisHost"
                :placeholder="'Redis ' + $t('m.Host')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="8" :xs="24">
            <el-form-item
              :label="'Redis ' + $t('m.Port')"
              label-width="80px"
              required
            >
              <el-input
                v-model="databaseConfig.redisPort"
                type="number"
                :placeholder="'Redis ' + $t('m.Port')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :md="8" :xs="24">
            <el-form-item
              :label="'Redis ' + $t('m.Password')"
              label-width="80px"
              required
            >
              <el-input
                v-model="databaseConfig.redisPassword"
                type="password"
                :placeholder="'Redis ' + $t('m.Password')"
              ></el-input>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <el-button
        type="primary"
        @click.native="saveDataBaseConfig"
        size="small"
        >{{ $t('m.Save') }}</el-button
      >
    </el-card>
  </div>
</template>

<script>
import api from '@/common/api';
import myMessage from '@/common/message';
import utils from '@/common/utils';
export default {
  name: 'SystemConfig',
  data() {
    return {
      init: false,
      saved: false,
      loadingBtnTest: false,
      smtp: {
        emailHost: 'smtp.example.com',
        emailPassword: '',
        emailPort: 587,
        emailBGImg: '',
        emailUsername: 'email@example.com',
        emailSsl: true,
      },
      websiteConfig: {},
      databaseConfig: {},
      dialogImageUrl: '',
      dialogVisible: false,
      disabled: false,
      carouselImgList: [],
      aiConfig: {
        enabled: false, apiUrl: '', apiKey: '', model: 'gpt-4o-mini', timeoutSeconds: 120,
        systemPrompt: '', validationPrompt: ''
      },
    };
  },
  mounted() {
    api.admin_getSMTPConfig().then((res) => {
      if (res.data.data) {
        this.smtp = res.data.data;
      } else {
        this.init = true;
        myMessage.warning('No STMP Config');
      }
    });

    api.getHomeCarousel().then((res) => {
      this.carouselImgList = res.data.data;
    });

    api
      .admin_getWebsiteConfig()
      .then((res) => {
        this.websiteConfig = res.data.data;
      })
        .catch(() => {}),
      api.admin_getProblemAIConfig().then((res) => { this.aiConfig = Object.assign(this.aiConfig, res.data.data || {}) }).catch(() => {}),
      api
        .admin_getDataBaseConfig()
        .then((res) => {
          this.databaseConfig = res.data.data;
        })
        .catch(() => {});
  },
  methods: {
    handleRemove(file, index = undefined) {
      let id = file.id;
      if (file.response != null) {
        id = file.response.data.id;
      }
      api.admin_deleteHomeCarousel(id).then((res) => {
        myMessage.success(this.$i18n.t('m.Delete_successfully'));
        if (index != undefined) {
          this.carouselImgList.splice(index, 1);
        }
      });
    },
    handlePictureCardPreview(file) {
      this.dialogImageUrl = file.url;
      this.dialogVisible = true;
    },
    handleDownload(file) {
      utils.downloadFile(file.url);
    },
    saveSMTPConfig() {
      api.admin_editSMTPConfig(this.smtp).then(
        (res) => {
          myMessage.success(this.$i18n.t('m.Update_Successfully'));
          this.saved = true;
        },
        () => {
          this.saved = false;
        }
      );
    },
    testSMTPConfig() {
      this.$prompt(this.$i18n.t('m.Please_input_your_email'), '', {
        inputPattern: /[\w!#$%&'*+/=?^_`{|}~-]+(?:\.[\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\w](?:[\w-]*[\w])?\.)+[\w](?:[\w-]*[\w])?/,
        inputErrorMessage: this.$i18n.t('m.Email_Check_Format'),
      })
        .then(({ value }) => {
          this.loadingBtnTest = true;
          api.admin_testSMTPConfig(value).then(
            (res) => {
              myMessage.success(this.$i18n.t('m.Send_successfully'));
              this.loadingBtnTest = false;
            },
            () => {
              this.loadingBtnTest = false;
            }
          );
        })
        .catch(() => {});
    },
    saveWebsiteConfig() {
      for (var key in this.websiteConfig) {
        if (key == 'register') {
          continue;
        } else {
          if (this.websiteConfig[key] == null || !this.websiteConfig[key].replace(/(^\s*)|(\s*$)/g, '')) {
            this.websiteConfig[key] = 'None';
          }
        }
      }
      api
        .admin_editWebsiteConfig(this.websiteConfig)
        .then((res) => {
          myMessage.success(this.$i18n.t('m.Update_Successfully'));
        })
        .catch(() => {});
    },
    saveDataBaseConfig() {
      api
        .admin_editDataBaseConfig(this.databaseConfig)
        .then((res) => {
          myMessage.success(this.$i18n.t('m.Update_Successfully'));
        })
        .catch(() => {});
    },
    saveAIConfig() {
      const timeoutSeconds = Number(this.aiConfig.timeoutSeconds);
      if (!Number.isFinite(timeoutSeconds) || timeoutSeconds < 1 || timeoutSeconds > 3600) {
        myMessage.warning('单请求超时必须在 1–3600 秒之间');
        return;
      }
      this.aiConfig.timeoutSeconds = Math.round(timeoutSeconds);
      api.admin_saveProblemAIConfig(this.aiConfig)
        .then(() => api.admin_getProblemAIConfig())
        .then((res) => {
          this.aiConfig = Object.assign(this.aiConfig, res.data.data || {});
          myMessage.success(this.$i18n.t('m.Update_Successfully'));
        });
    },
  },
};
</script>
