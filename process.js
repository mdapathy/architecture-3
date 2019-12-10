const fs = require('fs');

function parseCLArguments() {

    if (process.argv.length < 4) {
        console.error("Not enough parameters specified, required 2, got %d", process.argv.length - 2)
        process.exit();
    } else if (!fs.statSync(process.argv[2]).isDirectory()) {
        console.error(process.argv[2] + " is supposed to be a directory")
        process.exit();
    } else if (process.argv.length > 4) {
        console.log("Ignoring all parameters except " + process.argv[2] + " and " + process.argv[3])
    }

    var param = {}

    param.inputDir = process.argv[2]
    param.outputDir = process.argv[3]

    return param;
};

function getInputFiles(dir, filelist) {

    var path = path || require('path');
    var files = fs.readdirSync(dir),
        filelist = (filelist === null) ? [] : filelist;

    files.forEach(function(file) {
        if (fs.statSync(path.join(dir, file)).isDirectory()) {
            filelist = getInputFiles(path.join(dir, file), filelist);
        } else {
            filelist.push(path.join(dir, file));
        }

    });

    return filelist;
};

const parameters = parseCLArguments();
const inputfiles = getInputFiles(parameters.inputDir, []);


console.log("Total number of processed files:", inputfiles.length)