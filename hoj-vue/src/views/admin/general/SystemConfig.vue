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
      <div slot="header"><span class="panel-title home-title">{{ $t('m.Sys_AI_Config_Title') }}</span></div>
      <el-alert
        :title="$t('m.Sys_AI_Builtin_Title')"
        :description="$t('m.Sys_AI_Builtin_Desc')"
        type="info" :closable="false" show-icon style="margin-bottom:16px"
      />
      <el-alert v-if="aiConfigError" :title="aiConfigError" type="error" :closable="false" show-icon />
      <el-button v-if="aiConfigError && !aiConfigLoaded" type="text" @click="fetchAIConfig">{{ $t('m.ProbAI_Reload_Models') }}</el-button>
      <el-form v-loading="aiConfigLoading" label-width="120px" :model="aiConfig" :disabled="!aiConfigLoaded || aiConfigSaving">
        <el-row :gutter="20">
          <el-col :md="12" :xs="24"><el-form-item :label="$t('m.Sys_Enable_AI')"><el-switch v-model="aiConfig.enabled" /></el-form-item></el-col>
          <el-col :md="8" :xs="24"><el-form-item :label="$t('m.Sys_Timeout')"><el-input v-model.number="aiConfig.timeoutSeconds" type="number" :min="1" :max="3600" :step="30"><template slot="append">{{ $t('m.Sys_Seconds_Unit') }}</template></el-input><div style="color:#909399;font-size:12px;line-height:20px">{{ $t('m.Sys_Timeout_Hint') }}</div></el-form-item></el-col>
          <el-col :span="24">
            <div class="ai-model-hint">{{ $t('m.Sys_AI_Models_Hint') }}</div>
            <el-card v-for="item in aiConfig.models" :key="item.id" shadow="never" class="ai-model-card">
              <div slot="header" class="ai-model-heading">
                <el-radio :value="primaryModelId" :label="item.id" @input="setPrimaryModel">{{ $t('m.ProbAI_Primary_Model') }}</el-radio>
                <el-button type="text" :disabled="aiConfig.models.length === 1" @click="removeAIModel(item.id)">{{ $t('m.Delete') }}</el-button>
              </div>
              <el-row :gutter="20">
                <el-col :md="12" :xs="24"><el-form-item :label="$t('m.Sys_AI_Model_Name')" required><el-input v-model="item.name" maxlength="100" /></el-form-item></el-col>
                <el-col :md="12" :xs="24"><el-form-item :label="$t('m.Sys_Model')" required><el-input v-model="item.model" maxlength="100" /></el-form-item></el-col>
                <el-col :md="12" :xs="24">
                  <el-form-item :label="$t('m.Sys_AI_Protocol')" required>
                    <el-select v-model="item.protocol" style="width:100%">
                      <el-option v-for="option in aiProtocols" :key="option.value" :label="$t('m.' + option.labelKey)" :value="option.value" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :md="12" :xs="24"><el-form-item label="API Key" :required="aiConfig.enabled && !item.apiKeyConfigured"><el-input v-model="item.apiKey" maxlength="500" type="password" show-password autocomplete="new-password" :placeholder="item.apiKeyConfigured ? $t('m.Sys_AI_Key_Keep') : $t('m.Sys_AI_Key_Required')" /></el-form-item></el-col>
                <el-col :span="24">
                  <el-form-item :label="$t('m.Sys_Api_Url')" :required="aiConfig.enabled">
                    <el-input v-model="item.apiUrl" maxlength="500" :placeholder="aiProtocolOption(item.protocol).url" />
                    <div class="ai-protocol-hint">{{ $t('m.' + aiProtocolOption(item.protocol).hintKey) }}</div>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-card>
            <el-button size="small" plain icon="el-icon-plus" class="ai-add-model" @click="addAIModel">{{ $t('m.Sys_AI_Add_Model') }}</el-button>
          </el-col>
          <el-col :md="24" :xs="24"><el-form-item :label="$t('m.Sys_System_Prompt')"><el-input v-model="aiConfig.systemPrompt" type="textarea" :rows="8" /></el-form-item></el-col>
          <el-col :md="24" :xs="24"><el-form-item :label="$t('m.Sys_Validation_Prompt')"><el-input v-model="aiConfig.validationPrompt" type="textarea" :rows="10" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <el-button type="primary" size="small" :loading="aiConfigSaving" :disabled="!aiConfigLoaded || aiConfigLoading" @click="saveAIConfig">{{ $t('m.Save') }}</el-button>
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
      aiConfigLoading: true, aiConfigLoaded: false, aiConfigSaving: false, aiConfigError: '',
      aiProtocols: [
        { value: 'openai_chat', labelKey: 'Sys_AI_Protocol_OpenAI_Chat', url: 'https://api.openai.com/v1/chat/completions', hintKey: 'Sys_AI_Protocol_OpenAI_Chat_Hint' },
        { value: 'openai_responses', labelKey: 'Sys_AI_Protocol_OpenAI_Responses', url: 'https://api.openai.com/v1/responses', hintKey: 'Sys_AI_Protocol_OpenAI_Responses_Hint' },
        { value: 'gemini', labelKey: 'Sys_AI_Protocol_Gemini', url: 'https://generativelanguage.googleapis.com/v1beta', hintKey: 'Sys_AI_Protocol_Gemini_Hint' },
        { value: 'anthropic', labelKey: 'Sys_AI_Protocol_Anthropic', url: 'https://api.anthropic.com/v1/messages', hintKey: 'Sys_AI_Protocol_Anthropic_Hint' }
      ],
      aiConfig: {
        enabled: false, models: [], timeoutSeconds: 120,
        systemPrompt: '', validationPrompt: ''
      },
    };
  },
  computed: {
    primaryModelId() {
      const primary = this.aiConfig.models.find(item => item.primary);
      return primary ? primary.id : '';
    },
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
      this.fetchAIConfig(),
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
    async fetchAIConfig() {
      this.aiConfigLoading = true;
      this.aiConfigError = '';
      try {
        const res = await api.admin_getProblemAIConfig();
        this.loadAIConfig(res.data.data);
        this.aiConfigLoaded = true;
      } catch (e) {
        this.aiConfigLoaded = false;
        this.aiConfigError = this.$t('m.Sys_AI_Load_Failed');
      } finally {
        this.aiConfigLoading = false;
      }
    },
    loadAIConfig(config) {
      const data = config || {};
      this.aiConfig = {
        enabled: Boolean(data.enabled), timeoutSeconds: data.timeoutSeconds || 120,
        systemPrompt: data.systemPrompt || '', validationPrompt: data.validationPrompt || '',
        models: (data.models || []).map(item => Object.assign({}, item, { apiKey: '', protocol: item.protocol == null ? 'openai_chat' : item.protocol }))
      };
      if (!this.aiConfig.models.length) this.addAIModel();
    },
    aiProtocolOption(protocol) {
      return this.aiProtocols.find(option => option.value === protocol) || this.aiProtocols[0];
    },
    addAIModel() {
      const id = '10000000-1000-4000-8000-100000000000'.replace(/[018]/g, c =>
        (Number(c) ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> Number(c) / 4).toString(16));
      this.aiConfig.models.push({ id, name: '', apiUrl: '', apiKey: '', apiKeyConfigured: false,
        model: '', protocol: 'openai_chat', primary: this.aiConfig.models.length === 0 });
    },
    setPrimaryModel(id) {
      this.aiConfig.models.forEach(item => { item.primary = item.id === id; });
    },
    removeAIModel(id) {
      this.aiConfig.models = this.aiConfig.models.filter(item => item.id !== id);
      if (!this.aiConfig.models.some(item => item.primary) && this.aiConfig.models.length) {
        this.setPrimaryModel(this.aiConfig.models[0].id);
      }
    },
    async saveAIConfig() {
      if (!this.aiConfigLoaded || this.aiConfigLoading || this.aiConfigSaving) return;
      const timeoutSeconds = Number(this.aiConfig.timeoutSeconds);
      if (!Number.isFinite(timeoutSeconds) || timeoutSeconds < 1 || timeoutSeconds > 3600) {
        return myMessage.warning(this.$t('m.Sys_Timeout_Range_Warn'));
      }
      const models = this.aiConfig.models;
      if (!models.length || models.filter(item => item.primary).length !== 1) {
        return myMessage.warning(this.$t('m.Sys_AI_Primary_Required'));
      }
      for (const item of models) {
        item.name = String(item.name || '').trim();
        item.model = String(item.model || '').trim();
        item.apiUrl = String(item.apiUrl || '').trim();
        item.apiKey = String(item.apiKey || '').trim();
        if (!item.name || !item.model) return myMessage.warning(this.$t('m.Sys_AI_Model_Required'));
        if (!this.aiProtocols.some(option => option.value === item.protocol)) return myMessage.warning(this.$t('m.Sys_AI_Protocol_Invalid'));
        if (item.apiUrl) {
          try {
            const url = new URL(item.apiUrl);
            if (!['http:', 'https:'].includes(url.protocol) || url.username || url.password || url.hash) throw new Error('Invalid URL');
          } catch (e) { return myMessage.warning(this.$t('m.Sys_AI_Url_Invalid')); }
        }
        if (this.aiConfig.enabled && (!item.apiUrl || (!item.apiKey && !item.apiKeyConfigured))) {
          return myMessage.warning(this.$t('m.Sys_AI_Connection_Required'));
        }
      }
      this.aiConfig.timeoutSeconds = Math.round(timeoutSeconds);
      this.aiConfigSaving = true;
      this.aiConfigError = '';
      try {
        await api.admin_saveProblemAIConfig(this.aiConfig);
        this.aiConfig.models.forEach(item => {
          item.apiKeyConfigured = Boolean(item.apiKey || item.apiKeyConfigured);
          item.apiKey = '';
        });
        myMessage.success(this.$i18n.t('m.Update_Successfully'));
        await this.fetchAIConfig();
      } catch (e) {
        this.aiConfigError = this.$t('m.Sys_AI_Save_Failed');
      } finally {
        this.aiConfigSaving = false;
      }
    },
  },
};
</script>

<style scoped>
.ai-model-hint { margin-bottom: 12px; color: #606266; line-height: 1.6; }
.ai-model-card { margin-bottom: 12px; }
.ai-protocol-hint { color: #909399; font-size: 12px; line-height: 20px; margin-top: 6px; }
.ai-model-heading { display: flex; align-items: center; justify-content: space-between; }
.ai-add-model { margin-bottom: 20px; }
</style>
