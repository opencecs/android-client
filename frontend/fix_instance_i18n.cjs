const fs = require('fs');
let c = fs.readFileSync('src/components/instanceManagement.vue', 'utf8');

const replacements = [
  // Auth sync dialog
  [`placeholder="请输入用户名/手机号"`, `:placeholder="$t('instance.enterUsernameOrPhone')"`],
  [`label="密码" required`, `:label="$t('instance.password')" required`],
  [`placeholder="请输入密码"\r\n                        show-password\r\n                    ></el-input>\r\n                </el-form-item>\r\n                <el-form-item>\r\n                    <div style="display: flex; align-items: center; justify-content: space-between; width: 100%;">\r\n                        <el-checkbox v-model="syncAuthForm.saveCredentials">记住凭证</el-checkbox>\r\n                        <el-link type="primary" :underline="false" @click="openForgotPasswordDialog" style="font-size: 13px;">忘记密码</el-link>`,
   `:placeholder="$t('instance.enterPassword')"\r\n                        show-password\r\n                    ></el-input>\r\n                </el-form-item>\r\n                <el-form-item>\r\n                    <div style="display: flex; align-items: center; justify-content: space-between; width: 100%;">\r\n                        <el-checkbox v-model="syncAuthForm.saveCredentials">{{ $t('instance.rememberCredentials') }}</el-checkbox>\r\n                        <el-link type="primary" :underline="false" @click="openForgotPasswordDialog" style="font-size: 13px;">{{ $t('instance.forgotPassword') }}</el-link>`],
  [`<el-button @click="handleSyncAuthCancel">取消</el-button>`,
   `<el-button @click="handleSyncAuthCancel">{{ $t('instance.cancel') }}</el-button>`],
  [`{{ syncAuthLoading ? '登录中...' : '登录' }}`,
   `{{ syncAuthLoading ? $t('instance.loggingIn') : $t('instance.login') }}`],
  [`<el-button type="success" @click="openRegisterDialog">注册</el-button>`,
   `<el-button type="success" @click="openRegisterDialog">{{ $t('instance.register') }}</el-button>`],
  // Register form
  [`<el-form-item label="手机号" required>`, `<el-form-item :label="$t('instance.phone')" required>`],
  [`placeholder="请输入手机号"\r\n                        autocomplete="off"`,
   `:placeholder="$t('instance.enterPhone')"\r\n                        autocomplete="off"`],
  [`<el-form-item label="登录密码" required>`, `<el-form-item :label="$t('instance.loginPassword')" required>`],
  [`placeholder="请输入密码"\r\n                        show-password\r\n                    ></el-input>\r\n                </el-form-item>\r\n                <el-form-item label="确认密码" required>`,
   `:placeholder="$t('instance.enterPassword')"\r\n                        show-password\r\n                    ></el-input>\r\n                </el-form-item>\r\n                <el-form-item :label="$t('instance.confirmPassword')" required>`],
  [`placeholder="请再次输入密码"`, `:placeholder="$t('instance.enterPasswordAgain')"`],
  [`<el-form-item label="验证码" required>`, `<el-form-item :label="$t('instance.verificationCode')" required>`],
  [`placeholder="请输入验证码"`, `:placeholder="$t('instance.enterVerificationCode')"`],
  [`<el-button @click="handleRegisterCancel">取消</el-button>`,
   `<el-button @click="handleRegisterCancel">{{ $t('instance.cancel') }}</el-button>`],
  [`{{ registerLoading ? '注册中...' : '注册' }}`,
   `{{ registerLoading ? $t('instance.registering') : $t('instance.register') }}`],
];

let count = 0;
for (const [old, newStr] of replacements) {
  if (c.includes(old)) {
    c = c.replace(old, newStr);
    count++;
  } else {
    console.log('❌ Not found:', old.substring(0, 60));
  }
}
fs.writeFileSync('src/components/instanceManagement.vue', c, 'utf8');
console.log(`✅ Applied ${count} replacements`);
