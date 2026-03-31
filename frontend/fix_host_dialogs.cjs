const fs = require('fs');
let c = fs.readFileSync('src/components/HostManagement.vue', 'utf8');

c = c.replace(/'删除分组确认',\r\n      {\r\n        confirmButtonText: '确定',\r\n        cancelButtonText: '取消',/g, 
  `t('common.deleteGroupConfirm'),\r\n      {\r\n        confirmButtonText: t('common.confirm'),\r\n        cancelButtonText: t('common.cancel'),`);

c = c.replace(/'删除分组确认',\r\n    {\r\n      confirmButtonText: '确定',\r\n      cancelButtonText: '取消',/g, 
  `t('common.deleteGroupConfirm'),\r\n    {\r\n      confirmButtonText: t('common.confirm'),\r\n      cancelButtonText: t('common.cancel'),`);

c = c.replace(/ElMessageBox\.prompt\('请输入新的分组名称', '编辑分组', \{\r\n    confirmButtonText: '确定',\r\n    cancelButtonText: '取消',\r\n    inputValue: localGroupFilter\.value,\r\n    inputPattern: \/\\S+\/,\r\n    inputErrorMessage: '分组名称不能为空'\r\n  \}\)/g,
  `ElMessageBox.prompt(t('common.enterNewGroupName'), t('common.editGroup'), {\r\n    confirmButtonText: t('common.confirm'),\r\n    cancelButtonText: t('common.cancel'),\r\n    inputValue: localGroupFilter.value,\r\n    inputPattern: /\\S+/,\r\n    inputErrorMessage: t('common.groupNameCannotBeEmpty')\r\n  })`);

// Fallback if line endings are just \n
c = c.replace(/'删除分组确认',\n      {\n        confirmButtonText: '确定',\n        cancelButtonText: '取消',/g, 
  `t('common.deleteGroupConfirm'),\n      {\n        confirmButtonText: t('common.confirm'),\n        cancelButtonText: t('common.cancel'),`);

c = c.replace(/'删除分组确认',\n    {\n      confirmButtonText: '确定',\n      cancelButtonText: '取消',/g, 
  `t('common.deleteGroupConfirm'),\n    {\n      confirmButtonText: t('common.confirm'),\n      cancelButtonText: t('common.cancel'),`);

c = c.replace(/ElMessageBox\.prompt\('请输入新的分组名称', '编辑分组', \{\n    confirmButtonText: '确定',\n    cancelButtonText: '取消',\n    inputValue: localGroupFilter\.value,\n    inputPattern: \/\\S+\/,\n    inputErrorMessage: '分组名称不能为空'\n  \}\)/g,
  `ElMessageBox.prompt(t('common.enterNewGroupName'), t('common.editGroup'), {\n    confirmButtonText: t('common.confirm'),\n    cancelButtonText: t('common.cancel'),\n    inputValue: localGroupFilter.value,\n    inputPattern: /\\S+/,\n    inputErrorMessage: t('common.groupNameCannotBeEmpty')\n  })`);


fs.writeFileSync('src/components/HostManagement.vue', c, 'utf8');
console.log('Fixed HostManagement translations');
