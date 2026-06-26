const fs = require('fs');
const path = require('path');

const dir = path.join(__dirname, 'migrations', 'modules');
const files = fs.readdirSync(dir);

files.forEach(file => {
  if (file.endsWith('.sql')) {
    const filePath = path.join(dir, file);
    let content = fs.readFileSync(filePath, 'utf8');
    // replace CREATE TABLE that are not followed by IF NOT EXISTS
    content = content.replace(/CREATE TABLE\s+(?!IF NOT EXISTS)/ig, 'CREATE TABLE IF NOT EXISTS ');
    content = content.replace(/CREATE INDEX\s+(?!IF NOT EXISTS)/ig, 'CREATE INDEX IF NOT EXISTS ');
    fs.writeFileSync(filePath, content, 'utf8');
  }
});
console.log('Done!');
