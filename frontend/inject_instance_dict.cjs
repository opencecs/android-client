const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainJsContent = fs.readFileSync(mainJsPath, 'utf8');

const zhInstanceDict = `
    instance: {
      alipayTip: '请使用支付宝扫描二维码进行支付',
      authSync: '授权同步',
      confirmBuy: '确认购买',
      details: '详情',
      detailTitle: '实例详情',
      host: '归属主机',
      instanceIP: '云机实例 / IP',
      instances: '实例列表',
      loadingPackages: '正在加载套餐...',
      loadingQR: '正在生成支付二维码...',
      login: '登录',
      loginRequired: '需要登录后才能进行授权或操作',
      noInstance: '暂无实例',
      noPackages: '无套餐',
      purchaseRenew: '购买/续费授权',
      query: '查询',
      register: '注册',
      scanToPay: '扫码支付',
      searchPlaceholder: '搜索主机 IP 或云机实例',
      selectPackage: '选择套餐',
      statusExpired: '已过期',
      statusExpiring: '即将过期',
      statusNormal: '正常',
      subtotal: '小计',
      totalCount: '共有 {count} 台云机实例',
      username: '账号',
      validUntil: '有效期至'
    },`;

const enInstanceDict = `
    instance: {
      alipayTip: 'Please scan the QR code via Alipay to pay',
      authSync: 'Auth Sync',
      confirmBuy: 'Confirm Purchase',
      details: 'Details',
      detailTitle: 'Instance Details',
      host: 'Host',
      instanceIP: 'Instance / IP',
      instances: 'Instances',
      loadingPackages: 'Loading packages...',
      loadingQR: 'Generating Payment QR...',
      login: 'Login',
      loginRequired: 'Login required for authorization',
      noInstance: 'No Instances found',
      noPackages: 'No packages',
      purchaseRenew: 'Purchase/Renew Auth',
      query: 'Query',
      register: 'Register',
      scanToPay: 'Scan to Pay',
      searchPlaceholder: 'Search Host IP or Machine Name',
      selectPackage: 'Select Package',
      statusExpired: 'Expired',
      statusExpiring: 'Expiring Soon',
      statusNormal: 'Normal',
      subtotal: 'Subtotal',
      totalCount: 'Total {count} instances',
      username: 'Username',
      validUntil: 'Valid Until'
    },`;

let insertCount = 0;
// Insert it right before `backup: {`
// Wait, my backup block right now looks like `backup: { add: '新增'` or `backup: { add: 'Add'` because I just fixed it.
// To be safe, let's insert before `network: {` and let it be added before network.
mainJsContent = mainJsContent.replace(
  /\n\s*network: \{/g,
  function(match) {
    insertCount++;
    if (insertCount === 1) {
      return '\n' + zhInstanceDict + match;
    } else if (insertCount === 2) {
      return '\n' + enInstanceDict + match;
    }
    return match;
  }
);

fs.writeFileSync(mainJsPath, mainJsContent);
console.log('SUCCESS: Injected instance dicts');
