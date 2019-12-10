'use strict';

const fs = require('fs');

if (process.argv.length < 4) {
  console.error("Not enough parameters specified, required 2, got %d", process.argv.length - 2)
  process.exit();
} else if (!fs.statSync(process.argv[2]).isDirectory()) {
  console.error(process.argv[2] + " is supposed to be a directory")
  process.exit();
} else if (process.argv.length > 4) {
  console.log("Ignoring all parameters except " + process.argv[2] + " and " + process.argv[3])
}

const exec = async (file, taskLocation, resultLocation) => {
  let buf = new Buffer.alloc(2000);
  fs.stat(taskLocation+file, (err, data) => {
    fs.open(taskLocation+file, 'r', (err, fb) => {
      fs.read(fb, buf, 0, 2000, Math.max(0, data.size-2000), (err, bytesRead, buffer) => {
        let text = '';
        for(let i = 0; i < bytesRead; i++)
          text += String.fromCharCode(buffer[i]);
        fs.mkdirSync(resultLocation, {recursive: true});
        const arr = (text+'   ').split(/(?<=[.?!] )/g);
        fs.appendFile(resultLocation + file.split('.')[0] + '.res', arr[arr.length-2], (err) => {
          if(err) throw err;
        })
      })
    })
  });
}

fs.readdir(process.argv[2], (err, files) => {
  if(err) throw err;
  files.forEach(element => {
    setTimeout(() => (exec('/' + element, process.argv[2], process.argv[3])), 0);
  });
});
