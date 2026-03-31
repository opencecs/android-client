const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainJsContent = fs.readFileSync(mainJsPath, 'utf8');

// The block under `instance: {`
mainJsContent = mainJsContent.replace(/instanceIP: '云机实例 \/ IP'/g, "instanceIP: '实例 / IP'");
mainJsContent = mainJsContent.replace(/searchPlaceholder: '搜索主机 IP 或云机实例'/g, "searchPlaceholder: '搜索主机 IP 或实例'");
mainJsContent = mainJsContent.replace(/instancePrefix: '云机'/g, "instancePrefix: '实例'");

fs.writeFileSync(mainJsPath, mainJsContent);
console.log('SUCCESS: Changed 云机 to 实例');
