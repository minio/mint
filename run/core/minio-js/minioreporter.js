/*
 * MinIO Reporter for JSON formatted logging, (C) 2017-2023 MinIO, Inc.
 *
 * This file is part of MinIO Object Storage stack
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

var mocha = require('mocha');
module.exports = minioreporter;

function minioreporter(runner) {
  mocha.reporters.Base.call(this, runner);
   var self = this;

  runner.on('pass', function (test) {
    GenerateJsonEntry(test)
  });

  runner.on('fail', function (test, err) {
    GenerateJsonEntry(test, err)
  });

}

/**
 * Convert test result into a JSON object and print on the console.
 *
 * @api private
 * @param test, err
 */

function GenerateJsonEntry (test, err) {
  var res = test.title.split("_")
  var jsonEntry = {};

  jsonEntry.name = "minio-js"  
  
  if (res.length > 0 && res[0].length) {
    jsonEntry.function = res[0]
  }
  
  if (res.length > 1 && res[1].length) {
    jsonEntry.args = res[1]
  }

  jsonEntry.duration = test.duration
  
  if (res.length > 2 && res[2].length) {
    jsonEntry.alert = res[2]
  }

  if (err != null ) {
    jsonEntry.status = "FAIL"
    jsonEntry.error = err.stack.replace(/\n/g, " ").replace(/ +(?= )/g,'')
  } else {
    jsonEntry.status = "PASS"
  }

  process.stdout.write(JSON.stringify(jsonEntry) + "\n")
}
