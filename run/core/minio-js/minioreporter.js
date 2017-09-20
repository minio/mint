/*
 * Minio Reporter for JSON formatted logging, (C) 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

var mocha = require('mocha');
module.exports = minioreporter;

function minioreporter(runner) {
  mocha.reporters.Base.call(this, runner);
   var self = this;

  runner.on('pass', function (test) {
    var res = test.title.split("_");
    process.stdout.write(JSON.stringify({name: "minio-js", status: "PASS", function: res[0], args: res[1], alert: res[2], duration: test.duration}));
  });

  runner.on('fail', function (test) {
     test.err = errorJSON (test.error)
     var res = test.title.split("_");
    process.stdout.write(JSON.stringify({name: "minio-js", status: "FAIL", function: res[0], args: res[1], alert: res[2], duration: test.duration, error: test.err}));
  });

}


/**
 * Transform `error` into a JSON object.
 *
 * @api private
 * @param {Error} err
 * @return {Object}
 */
function errorJSON (err) {

  if (err == null )
 return "";
  var res = {};
  Object.getOwnPropertyNames(err).forEach(function (key) {
    res[key] = err[key];
  }, err);
  return res;
}
