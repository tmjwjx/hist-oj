<template>
  <div>
    <template v-if="!mobileNar">
      <div id="header">
        <el-menu
          :default-active="activeMenuName"
          mode="horizontal"
          router
          active-text-color="#2196f3"
          text-color="#495060"
        >
          <div class="logo">
            <el-tooltip
              :content="$t('m.Click_To_Change_Web_Language')"
              placement="bottom"
              effect="dark"
            >
              <el-image
                style="width: 139px; height: 50px"
                :src="imgUrl"
                fit="scale-down"
                @click="changeWebLanguage"
              ></el-image>
            </el-tooltip>
          </div>
          <template v-if="mode == 'defalut'">
            <el-menu-item index="/home"
              ><i class="el-icon-s-home"></i>{{ $t('m.NavBar_Home') }}</el-menu-item
            >
            <el-menu-item index="/problem"
              ><i class="el-icon-s-grid"></i
              >{{ $t('m.NavBar_Problem') }}</el-menu-item
            >
            <el-menu-item index="/training"
              ><i class="el-icon-s-claim"></i
              >{{ $t('m.NavBar_Training') }}</el-menu-item
            >
            <el-menu-item index="/contest"
              ><i class="el-icon-trophy"></i
              >{{ $t('m.NavBar_Contest') }}</el-menu-item
            >
            <el-menu-item index="/status"
              ><i class="el-icon-s-marketing"></i
              >{{ $t('m.NavBar_Status') }}</el-menu-item
            >
            <el-submenu index="rank">
              <template slot="title"
                ><i class="el-icon-s-data"></i>{{ $t('m.NavBar_Rank') }}</template
              >
              <el-menu-item index="/acm-rank">{{
                $t('m.NavBar_ACM_Rank')
              }}</el-menu-item>
              <el-menu-item index="/oi-rank">{{
                $t('m.NavBar_OI_Rank')
              }}</el-menu-item>
              <el-menu-item index="/rating-rank">
                <i class="el-icon-trophy"></i> Rating 排名
              </el-menu-item>
            </el-submenu>
            <el-menu-item index="/discussion"
              v-if="websiteConfig.openPublicDiscussion"
              ><i class="el-icon-s-comment"></i
              >{{ $t('m.NavBar_Discussion') }}</el-menu-item
            >
            <el-menu-item index="/group"
              ><i
                class="fa fa-users navbar-icon"
              ></i
              >{{ $t('m.NavBar_Group') }}</el-menu-item
            >
            <el-menu-item index="/toolbox"
              ><i class="fa fa-briefcase"></i>工具箱</el-menu-item
            >
            <el-menu-item index="/introduction"
              ><i class="el-icon-document"></i>编译环境</el-menu-item
            >
            <el-menu-item index="/about-us"
              ><i class="el-icon-user"></i>关于我们</el-menu-item
            >
        </template>
        <template v-else-if="mode == 'training'">
          <el-menu-item index="/home"
              ><i class="el-icon-s-home"></i>{{ $t('m.NavBar_Back_Home') }}</el-menu-item
            >
            <template v-if="$route.params.groupID">
              <el-menu-item :index="'/group/' + $route.params.groupID"
              ><i
                class="fa fa-users navbar-icon"
              ></i
              >{{ $t('m.NavBar_Group_Home') }}</el-menu-item>
            </template>
            <el-menu-item :index="getTrainingHomePath()"
              ><i class="el-icon-s-claim"></i>{{ $t('m.NavBar_Training_Home') }}</el-menu-item
            >
            <el-menu-item :index="getTrainingProblemListPath()"
              ><i class="fa fa-list navbar-icon"></i>{{ $t('m.Problem_List') }}</el-menu-item
            >
        </template>
        <template v-else-if="mode == 'contest'">
          <el-menu-item index="/home"
              ><i class="el-icon-s-home"></i>{{ $t('m.NavBar_Back_Home') }}</el-menu-item
            >
            <el-menu-item :index="'/contest/' + $route.params.contestID"
              ><i class="el-icon-trophy"></i>{{ $t('m.NavBar_Contest_Home') }}</el-menu-item
            >
            <el-menu-item :index="'/contest/' + $route.params.contestID + '/problems'"
              ><i class="fa fa-list navbar-icon"></i>{{ $t('m.Problem_List') }}</el-menu-item
            >
            <el-menu-item :index="'/contest/' + $route.params.contestID + '/submissions?onlyMine=true'"
              ><i class="el-icon-menu"></i>{{ $t('m.NavBar_Contest_Own_Submission') }}</el-menu-item
            >
            <el-menu-item :index="'/contest/' + $route.params.contestID + '/rank'"
              ><i class="fa fa-bar-chart navbar-icon"></i>{{ $t('m.NavBar_Contest_Rank') }}</el-menu-item
            >
        </template>
        <template v-else-if="mode == 'group'">
          <el-menu-item index="/home"
              ><i class="el-icon-s-home"></i>{{ $t('m.NavBar_Back_Home') }}</el-menu-item
            >
            <template v-if="$route.params.groupID">
              <el-menu-item :index="'/group/' + $route.params.groupID"
              ><i
                class="fa fa-users navbar-icon"
              ></i
              >{{ $t('m.NavBar_Group_Home') }}</el-menu-item>
            </template>
            <el-menu-item :index="'/group/' + $route.params.groupID + '/problem'"
              ><i class="fa fa-list navbar-icon"></i>{{ $t('m.Problem_List') }}</el-menu-item
            >
        </template>

          <template v-if="!isAuthenticated">
            <div class="btn-menu">
              <el-button 
                type="primary" 
                size="medium" 
                round
                @click="handleBtnClick('Login')"
                >{{ $t('m.NavBar_Login') }}
              </el-button>
              <el-button
                v-if="websiteConfig.register"
                size="medium"
                round
                @click="handleBtnClick('Register')"
                style="margin-left: 5px"
                >{{ $t('m.NavBar_Register') }}
              </el-button>
            </div>
          </template>
          <template v-else>
            <el-dropdown
              class="drop-menu"
              @command="handleRoute"
              placement="bottom"
              trigger="hover"
            >
              <span class="el-dropdown-link" :style="getUsernameStyle()">
                {{ userInfo.username }}<i class="el-icon-caret-bottom"></i>
              </span>

              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="/user-home">{{
                  $t('m.NavBar_UserHome')
                }}</el-dropdown-item>
                <el-dropdown-item command="/status?onlyMine=true">{{
                  $t('m.NavBar_Submissions')
                }}</el-dropdown-item>
                <el-dropdown-item command="/setting">{{
                  $t('m.NavBar_Setting')
                }}</el-dropdown-item>
                <el-dropdown-item v-if="isAdminRole" command="/admin">{{
                  $t('m.NavBar_Management')
                }}</el-dropdown-item>
                <el-dropdown-item divided command="/logout">{{
                  $t('m.NavBar_Logout')
                }}</el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
            <avatar
              :username="userInfo.username"
              :inline="true"
              :size="30"
              color="#FFF"
              :src="avatar"
              class="drop-avatar"
            ></avatar>
            <el-dropdown
              class="drop-msg"
              @command="handleRoute"
              placement="bottom"
            >
              <span class="el-dropdown-link">
                <i class="el-icon-message-solid"></i>
                <svg
                  v-if="
                    unreadMessage.comment > 0 ||
                      unreadMessage.reply > 0 ||
                      unreadMessage.like > 0 ||
                      unreadMessage.sys > 0 ||
                      unreadMessage.mine > 0
                  "
                  width="10"
                  height="10"
                  style="vertical-align: top;margin-left: -11px;margin-top: 3px;"
                >
                  <circle cx="5" cy="5" r="5" style="fill: red;"></circle>
                </svg>
              </span>

              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="/message/discuss">
                  <span>{{ $t('m.DiscussMsg') }}</span>
                  <span class="drop-msg-count" v-if="unreadMessage.comment > 0">
                    <MsgSvg :total="unreadMessage.comment"></MsgSvg>
                  </span>
                </el-dropdown-item>
                <el-dropdown-item command="/message/reply">
                  <span>{{ $t('m.ReplyMsg') }}</span>
                  <span class="drop-msg-count" v-if="unreadMessage.reply > 0">
                    <MsgSvg :total="unreadMessage.reply"></MsgSvg>
                  </span>
                </el-dropdown-item>
                <el-dropdown-item command="/message/like">
                  <span>{{ $t('m.LikeMsg') }}</span>
                  <span class="drop-msg-count" v-if="unreadMessage.like > 0">
                    <MsgSvg :total="unreadMessage.like"></MsgSvg>
                  </span>
                </el-dropdown-item>
                <el-dropdown-item command="/message/sys">
                  <span>{{ $t('m.SysMsg') }}</span>
                  <span class="drop-msg-count" v-if="unreadMessage.sys > 0">
                    <MsgSvg :total="unreadMessage.sys"></MsgSvg>
                  </span>
                </el-dropdown-item>
                <el-dropdown-item command="/message/mine">
                  <span>{{ $t('m.MineMsg') }}</span>
                  <span class="drop-msg-count" v-if="unreadMessage.mine > 0">
                    <MsgSvg :total="unreadMessage.mine"></MsgSvg>
                  </span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-menu>
      </div>
      <div id="header-hidden" v-show="isScrolled">
      </div>
    </template>
    <template v-else>
      <div style="top:0px;left:0px;">
      <mu-appbar class="mobile-nav" color="primary">
        <mu-button icon slot="left" @click="opendrawer = !opendrawer">
          <i class="el-icon-s-unfold"></i>
        </mu-button>
        <el-tooltip
            :content="$t('m.Click_To_Change_Web_Language')"
            placement="bottom"
            effect="dark"
          >
          <span @click="changeWebLanguage">
          {{
            websiteConfig.shortName ? websiteConfig.shortName : 'OJ'
          }}
          </span>
        </el-tooltip>
        <mu-button
          flat
          slot="right"
          @click="handleBtnClick('Login')"
          v-show="!isAuthenticated"
          >{{ $t('m.NavBar_Login') }}</mu-button
        >
        <mu-button
          flat
          slot="right"
          @click="handleBtnClick('Register')"
          v-show="!isAuthenticated && websiteConfig.register"
          >{{ $t('m.NavBar_Register') }}</mu-button
        >

        <mu-menu slot="right" v-show="isAuthenticated" :open.sync="openmsgmenu">
          <mu-button flat>
            <mu-icon value=":el-icon-message-solid" size="24"></mu-icon>
            <svg
              v-if="
                unreadMessage.comment > 0 ||
                  unreadMessage.reply > 0 ||
                  unreadMessage.like > 0 ||
                  unreadMessage.sys > 0 ||
                  unreadMessage.mine > 0
              "
              width="10"
              height="10"
              style="margin-left: -11px;margin-top: -13px;"
            >
              <circle cx="5" cy="5" r="5" style="fill: red;"></circle>
            </svg>
          </mu-button>
          <mu-list slot="content" @change="handleCommand">
            <mu-list-item button value="/message/discuss">
              <mu-list-item-content>
                <mu-list-item-title>
                  {{ $t('m.DiscussMsg') }}
                  <span class="drop-msg-count" v-if="unreadMessage.comment > 0">
                    <MsgSvg :total="unreadMessage.comment"></MsgSvg>
                  </span>
                </mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
            <mu-divider></mu-divider>
            <mu-list-item button value="/message/reply">
              <mu-list-item-content>
                <mu-list-item-title>
                  {{ $t('m.ReplyMsg') }}
                  <span class="drop-msg-count" v-if="unreadMessage.reply > 0">
                    <MsgSvg :total="unreadMessage.reply"></MsgSvg>
                  </span>
                </mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
            <mu-divider></mu-divider>
            <mu-list-item button value="/message/like">
              <mu-list-item-content>
                <mu-list-item-title>
                  {{ $t('m.LikeMsg') }}
                  <span class="drop-msg-count" v-if="unreadMessage.like > 0">
                    <MsgSvg :total="unreadMessage.like"></MsgSvg>
                  </span>
                </mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
            <mu-divider></mu-divider>
            <mu-list-item button value="/message/sys">
              <mu-list-item-content>
                <mu-list-item-title>
                  {{ $t('m.SysMsg') }}
                  <span class="drop-msg-count" v-if="unreadMessage.sys > 0">
                    <MsgSvg :total="unreadMessage.sys"></MsgSvg>
                  </span>
                </mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
            <mu-divider></mu-divider>

            <mu-list-item button value="/message/mine">
              <mu-list-item-content>
                <mu-list-item-title>
                  {{ $t('m.MineMsg') }}
                  <span class="drop-msg-count" v-if="unreadMessage.mine > 0">
                    <MsgSvg :total="unreadMessage.mine"></MsgSvg>
                  </span>
                </mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
          </mu-list>
        </mu-menu>

        <mu-menu
          slot="right"
          v-if="isAuthenticated"
          :open.sync="openusermenu"
        >
          <mu-button flat>
            <avatar
              :username="userInfo.username"
              :inline="true"
              :size="30"
              color="#FFF"
              :src="avatar"
              :title="userInfo.username"
            ></avatar>
            <i class="el-icon-caret-bottom"></i>
          </mu-button>
          <mu-list slot="content" @change="handleCommand">
            <mu-list-item button value="/user-home">
              <mu-list-item-content>
                <mu-list-item-title>{{
                  $t('m.NavBar_UserHome')
                }}</mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
            <mu-divider></mu-divider>
            <mu-list-item button value="/status?onlyMine=true">
              <mu-list-item-content>
                <mu-list-item-title>{{
                  $t('m.NavBar_Submissions')
                }}</mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
            <mu-divider></mu-divider>
            <mu-list-item button value="/setting">
              <mu-list-item-content>
                <mu-list-item-title>{{
                  $t('m.NavBar_Setting')
                }}</mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
            <mu-divider></mu-divider>
            <mu-list-item button value="/admin" v-show="isAdminRole">
              <mu-list-item-content>
                <mu-list-item-title>{{
                  $t('m.NavBar_Management')
                }}</mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
            <mu-divider></mu-divider>

            <mu-list-item button value="/logout">
              <mu-list-item-content>
                <mu-list-item-title>{{
                  $t('m.NavBar_Logout')
                }}</mu-list-item-title>
              </mu-list-item-content>
            </mu-list-item>
          </mu-list>
        </mu-menu>
      </mu-appbar>

      <mu-appbar style="width: 100%;">
        <!--占位，刚好占领导航栏的高度-->
      </mu-appbar>

      <mu-drawer :open.sync="opendrawer" :docked="false" :right="false">
        <mu-list toggle-nested>
          <mu-list-item
            button
            to="/home"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":el-icon-s-home" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>{{ $t('m.NavBar_Home') }}</mu-list-item-title>
          </mu-list-item>

          <mu-list-item
            button
            to="/problem"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":el-icon-s-grid" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>{{
              $t('m.NavBar_Problem')
            }}</mu-list-item-title>
          </mu-list-item>

          <mu-list-item
            button
            to="/training"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":el-icon-s-claim" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>{{
              $t('m.NavBar_Training')
            }}</mu-list-item-title>
          </mu-list-item>

          <mu-list-item
            button
            to="/contest"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":el-icon-trophy" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>{{
              $t('m.NavBar_Contest')
            }}</mu-list-item-title>
          </mu-list-item>

          <mu-list-item
            button
            to="/status"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":el-icon-s-marketing" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>{{ $t('m.NavBar_Status') }}</mu-list-item-title>
          </mu-list-item>

          <mu-list-item
            button
            :ripple="false"
            nested
            :open="openSideMenu === 'rank'"
            @toggle-nested="openSideMenu = arguments[0] ? 'rank' : ''"
          >
            <mu-list-item-action>
              <mu-icon value=":el-icon-s-data" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>{{ $t('m.NavBar_Rank') }}</mu-list-item-title>
            <mu-list-item-action>
              <mu-icon
                class="toggle-icon"
                size="24"
                value=":el-icon-arrow-down"
              ></mu-icon>
            </mu-list-item-action>
            <mu-list-item
              button
              :ripple="false"
              slot="nested"
              to="/acm-rank"
              @click="opendrawer = !opendrawer"
              active-class="mobile-menu-active"
            >
              <mu-list-item-title>{{
                $t('m.NavBar_ACM_Rank')
              }}</mu-list-item-title>
            </mu-list-item>
            <mu-list-item
              button
              :ripple="false"
              slot="nested"
              to="/oi-rank"
              @click="opendrawer = !opendrawer"
              active-class="mobile-menu-active"
            >
              <mu-list-item-title>{{
                $t('m.NavBar_OI_Rank')
              }}</mu-list-item-title>
            </mu-list-item>
          </mu-list-item>

          <mu-list-item
            v-if="websiteConfig.openPublicDiscussion"
            button
            to="/discussion"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":fa fa-comments" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>{{
              $t('m.NavBar_Discussion')
            }}</mu-list-item-title>
          </mu-list-item>

          <mu-list-item
            button
            to="/group"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":fa fa-users" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>{{ $t('m.NavBar_Group') }}</mu-list-item-title>
          </mu-list-item>

          <mu-list-item
            button
            :ripple="false"
            to="/registration"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":el-icon-s-claim" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>赛事报名系统</mu-list-item-title>
          </mu-list-item>

          <mu-list-item
            button
            :ripple="false"
            to="/introduction"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":el-icon-document" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>编译环境</mu-list-item-title>
          </mu-list-item>

          <mu-list-item
            button
            :ripple="false"
            to="/about-us"
            @click="opendrawer = !opendrawer"
            active-class="mobile-menu-active"
          >
            <mu-list-item-action>
              <mu-icon value=":el-icon-user" size="24"></mu-icon>
            </mu-list-item-action>
            <mu-list-item-title>关于我们</mu-list-item-title>
          </mu-list-item>
        </mu-list>
      </mu-drawer>
    </div>
    </template>
    
    <el-dialog
      :visible.sync="modalVisible"
      width="370px"
      class="dialog"
      :title="title"
      :close-on-click-modal="false"
    >
      <component :is="modalStatus.mode" v-if="modalVisible"></component>
      <div slot="footer" style="display: none"></div>
    </el-dialog>
  </div>
</template>
<script>
import Login from '@/components/oj/common/Login';
import Register from '@/components/oj/common/Register';
import ResetPwd from '@/components/oj/common/ResetPassword';
import MsgSvg from '@/components/oj/msg/msgSvg';
import { mapGetters, mapActions } from 'vuex';
import Avatar from 'vue-avatar';
import api from '@/common/api';
import ratingApi from '@/common/rating-api';
import { getRatingColor } from '@/common/rating-utils';
export default {
  components: {
    Login,
    Register,
    ResetPwd,
    Avatar,
    MsgSvg,
  },
  created(){
    this.page_width();
    window.onresize = () => {
      this.page_width();
      this.setHiddenHeaderHeight();
    };
  },
  mounted() {
    // 使用 $nextTick 确保 DOM 渲染完成后再设置模式
    this.$nextTick(() => {
      this.switchMode();
      this.setHiddenHeaderHeight();
    });
    if (this.isAuthenticated) {
      this.getUnreadMsgCount();
      this.msgTimer = setInterval(() => {
        this.getUnreadMsgCount();
      }, 120 * 1000);
      // 加载用户 Rating 颜色
      this.loadUserRatingColor();
    }
  },
  beforeDestroy() {
    clearInterval(this.msgTimer);
  },
  data() {
    return {
      mode:'defalut',
      centerDialogVisible: false,
      mobileNar: false,
      opendrawer: false,
      openusermenu: false,
      openmsgmenu: false,
      openSideMenu: '',
      imgUrl: require('@/assets/logo.png'),
      avatarStyle:
        'display: inline-flex;width: 30px;height: 30px;border-radius: 50%;align-items: center;justify-content: center;text-align: center;user-select: none;',
      userRatingColor: null,
      ratingLoaded: false,
      lastUnreadCount: 0, // 记录上次的未读消息总数，用于检测新消息
    };
  },
  methods: {
    ...mapActions(['changeModalStatus']),
    page_width() {
      let screenWidth = window.screen.width;
      if (screenWidth < 992) {
        this.mobileNar = true;
      } else {
        this.mobileNar = false;
      }
    },
    handleBtnClick(mode) {
      this.changeModalStatus({
        mode,
        visible: true,
      });
    },
    handleRoute(route) {
      //电脑端导航栏路由跳转事件
      if (route && route.split('/')[1] != 'admin') {
        this.$router.push(route);
      } else {
        window.open('/admin/');
      }
    },
    handleCommand(route) {
      // 移动端导航栏路由跳转事件
      this.openusermenu = false;
      this.openmsgmenu = false;
      if (route && route.split('/')[1] != 'admin') {
        this.$router.push(route);
      } else {
        window.open('/admin/');
      }
    },
    getUnreadMsgCount() {
      api.getUnreadMsgCount().then((res) => {
        let data = res.data.data;
        this.$store.dispatch('updateUnreadMessageCount', data);
        let sumMsg =
          data.comment + data.reply + data.like + data.mine + data.sys;

        // 只有当未读消息数量增加时才弹窗通知（避免重复通知已读的消息）
        if (sumMsg > 0 && sumMsg > this.lastUnreadCount) {
          if (this.webLanguage == 'zh-CN') {
            this.$notify.info({
              title: '未读消息',
              message:
                '亲爱的【' +
                this.userInfo.username +
                '】，您有最新的' +
                sumMsg +
                '条未读消息，请注意查看！',
              position: 'bottom-right',
              duration: 5000,
            });
          } else {
            this.$notify.info({
              title: 'Unread Message',
              message:
                'Dear【' +
                this.userInfo.username +
                '】, you have the latest ' +
                sumMsg +
                ' unread messages. Please check them!',
              position: 'bottom-right',
              duration: 5000,
            });
          }
        }
        // 更新上次的未读消息数
        this.lastUnreadCount = sumMsg;
      });
    },
    changeWebLanguage() {
      this.$store.commit('changeWebLanguage', { language: this.webLanguage == 'zh-CN' ? 'en-US' : 'zh-CN' });
    },
    setHiddenHeaderHeight(){
      if(!this.mobileNar){
        try {
          let headerHeight = document.getElementById('header').offsetHeight;
          document.getElementById('header-hidden').setAttribute('style','height:'+ headerHeight + 'px')
        } catch (e) {}
      }
    },
    switchMode(){
      if(this.$route.meta.fullScreenSource){
        this.mode = this.$route.meta.fullScreenSource;
      }else{
        this.mode = 'defalut';
      }
    },
    getTrainingHomePath(){
      let tid = this.$route.params.trainingID
      let gid = this.$route.params.groupID
      if(gid){
        return `/group/${gid}/training/${tid}`;
      }else{
        return `/training/${tid}`;
      }
    },
    getTrainingProblemListPath(){
      let tid = this.$route.params.trainingID
      let gid = this.$route.params.groupID
      if(gid){
        return `/group/${gid}/training/${tid}/problems`;
      }else{
        return `/training/${tid}/problems`;
      }
    },
    // 加载用户 Rating 颜色
    loadUserRatingColor() {
      if (!this.userInfo || !this.userInfo.uid) {
        this.ratingLoaded = true;
        // 如果没有用户信息，保持蓝色（不设置 userRatingColor）
        this.userRatingColor = null;
        return;
      }
      ratingApi.getUserRating(this.userInfo.uid).then(
        (data) => {
          // 优先使用 API 返回的 color
          if (data && typeof data.color === 'string' && data.color.trim() !== '') {
            this.userRatingColor = data.color;
          } 
          // 如果 API 返回了有效的 rating 值，使用前端函数计算颜色
          else if (data && typeof data.rating === 'number' && !isNaN(data.rating)) {
            this.userRatingColor = getRatingColor(data.rating);
          } 
          // 如果数据无效，保持蓝色
          else {
            this.userRatingColor = null;
          }
          this.ratingLoaded = true;
        },
        (err) => {
          console.error('加载用户 Rating 颜色失败:', err);
          // API 调用失败时，保持蓝色（不设置 userRatingColor）
          this.userRatingColor = null;
          this.ratingLoaded = true;
        }
      );
    },
    // 获取用户名样式
    getUsernameStyle() {
      // 如果还没有加载完成，显示蓝色
      if (!this.ratingLoaded) {
        return { 
          color: '#409eff',
          fontWeight: 'normal'
        };
      }
      // 加载完成后，如果有 rating 颜色就使用，否则保持蓝色
      return {
        color: this.userRatingColor || '#409eff',
        fontWeight: this.userRatingColor ? 'bold' : 'normal',
        transition: this.userRatingColor ? 'color 0.3s ease' : 'none'
      };
    }
  },
  computed: {
    ...mapGetters([
      'modalStatus',
      'userInfo',
      'isAuthenticated',
      'isAdminRole',
      'token',
      'websiteConfig',
      'unreadMessage',
      'webLanguage',
    ]),
    avatar() {
      return this.$store.getters.userInfo.avatar;
    },
    activeMenuName() {
      if (this.$route.path.split('/')[1] == 'submission-detail') {
        return '/status';
      } else if (this.$route.path.split('/')[1] == 'discussion-detail') {
        return '/discussion';
      }
      return '/' + this.$route.path.split('/')[1];
    },
    modalVisible: {
      get() {
        return this.modalStatus.visible;
      },
      set(value) {
        this.changeModalStatus({ visible: value });
      },
    },
    title: {
      get() {
        let ojName = this.websiteConfig.shortName
          ? this.websiteConfig.shortName
          : 'OJ';
        if (this.modalStatus.mode == 'ResetPwd') {
          return this.$i18n.t('m.Dialog_Reset_Password') + ' - ' + ojName;
        } else {
          return (
            this.$i18n.t('m.Dialog_' + this.modalStatus.mode) + ' - ' + ojName
          );
        }
      },
    },
  },
  watch: {
    isAuthenticated() {
      if (this.isAuthenticated) {
        if (this.msgTimer) {
          clearInterval(this.msgTimer);
        }
        this.getUnreadMsgCount();
        this.msgTimer = setInterval(() => {
          this.getUnreadMsgCount();
        }, 120 * 1000);
        // 加载用户 Rating 颜色
        this.loadUserRatingColor();
      } else {
        clearInterval(this.msgTimer);
        // 用户登出时重置 Rating 颜色
        this.userRatingColor = null;
        this.ratingLoaded = false;
      }
    },
    // 监听用户 UID 变化，确保切换用户时重新加载颜色
    'userInfo.uid'(newUid, oldUid) {
      if (this.isAuthenticated && newUid && newUid !== oldUid) {
        // 重置状态
        this.ratingLoaded = false;
        this.userRatingColor = null;
        // 重新加载颜色
        this.loadUserRatingColor();
      }
    },
    $route(){
      this.switchMode();
    }
  },
};
</script>
<style scoped>
#header {
  min-width: 300px;
  position: fixed;
  top: 0;
  left: 0;
  height: auto;
  width: 100%;
  z-index: 2000;
  background-color: #fff;
  box-shadow: 0 1px 5px 0 rgba(0, 0, 0, 0.1);
}
.mobile-nav {
  position: fixed;
  left: 0px;
  top: 0px;
  z-index: 2500;
  height: auto;
  width: 100%;
}

#drawer {
  position: fixed;
  left: 0px;
  bottom: 0px;
  z-index: 1000;
  width: 100%;
  box-shadow: 00px 0px 00px rgb(255, 255, 255), 0px 0px 10px rgb(255, 255, 255),
    0px 0px 0px rgb(255, 255, 255), 1px 1px 0px rgb(218, 218, 218);
}

.logo {
  cursor: pointer;
  margin-left: 2%;
  margin-right: 2%;
  float: left;
  width: 139px;
  height: 42px;
  margin-top: 5px;
}
.el-dropdown-link {
  cursor: pointer;
  /* color 由 getUsernameStyle() 动态设置，不再使用固定颜色 */
}
.el-icon-arrow-down {
  font-size: 18px;
}
.drop-menu {
  float: right;
  margin-right: 30px;
  position: relative;
  font-weight: 500;
  right: 10px;
  margin-top: 18px;
  font-size: 18px;
}
.drop-avatar {
  float: right;
  margin-right: 15px;
  position: relative;
  margin-top: 16px;
}
.drop-msg {
  float: right;
  font-size: 25px;
  margin-right: 10px;
  position: relative;
  margin-top: 13px;
}
.drop-msg-count {
  margin-left: 2px;
}
.btn-menu {
  font-size: 16px;
  float: right;
  margin-right: 10px;
  margin-top: 12px;
}
/deep/ .el-dialog {
  border-radius: 10px !important;
  text-align: center;
}
/deep/ .el-dialog__header .el-dialog__title {
  font-size: 22px;
  font-weight: 600;
  font-family: Arial, Helvetica, sans-serif;
  line-height: 1em;
  color: #4e4e4e;
}
.el-submenu__title i {
  color: #495060 !important;
}
.el-menu-item {
  padding: 0 13px;
}
.el-menu-item:hover, .el-menu .el-menu-item:hover{
  border-bottom: 2px solid #2474b5 !important;
}
.el-menu .el-menu-item:hover, 
.el-menu .el-menu-item:hover i,
.el-submenu .el-submenu__title:hover,
.el-submenu .el-submenu__title:hover i{
  outline: 0 !important;
  color: #2E95FB !important;
  background: linear-gradient(270deg, #F2F7FC 0%, #FEFEFE 100%)!important;
  transition: all .2s ease;
}
.el-menu .el-menu-item.is-active, 
.el-menu .el-menu-item.is-active i,
.el-submenu.is-active,
.el-submenu.is-active i
{
  color: #2E95FB !important;
  background: linear-gradient(270deg, #F2F7FC 0%, #FEFEFE 100%)!important;
  transition: all .2s ease;
}
.el-menu--horizontal .el-menu .el-menu-item:hover, 
.el-submenu /deep/.el-submenu__title:hover {
  color: #2E95FB !important;
  background: linear-gradient(270deg, #F2F7FC 0%, #FEFEFE 100%)!important;
}
.el-menu-item i {
  color: #495060;
}
.is-active .el-submenu__title i,
.is-active {
  color: #2196f3 !important;
}
.el-menu-item.is-active i {
  color: #2196f3 !important;
}
.navbar-icon{
  margin-right: 5px !important;
  width: 24px !important;
  text-align: center !important;
}
</style>
