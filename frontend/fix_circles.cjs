const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainText = fs.readFileSync(mainJsPath, 'utf8');

// Fix stretch issue in the Flex container for the guide circles
mainText = mainText.replace(/border-radius: 50%; line-height: 20px;"/g, 'height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;"');

// Also make the li have align-items: flex-start so they don't stretch
mainText = mainText.replace(/<li style="margin-bottom: 12px; display: flex;">/g, '<li style="margin-bottom: 12px; display: flex; align-items: flex-start;">');

fs.writeFileSync(mainJsPath, mainText);
console.log('SUCCESS: Fixed stretched circles');
