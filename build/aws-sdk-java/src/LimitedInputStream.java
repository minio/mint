/*
*  Mint, (C) 2017-2023 MinIO, Inc.
*
*  This file is part of MinIO Object Storage stack
*
*  This program is free software: you can redistribute it and/or modify
*  it under the terms of the GNU Affero General Public License as published by
*  the Free Software Foundation, either version 3 of the License, or
*  (at your option) any later version.
*
*  This program is distributed in the hope that it will be useful
*  but WITHOUT ANY WARRANTY; without even the implied warranty of
*  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
*  GNU Affero General Public License for more details.
*
*  You should have received a copy of the GNU Affero General Public License
*  along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package io.minio.awssdk.tests;

import java.io.*;

// LimitedInputStream wraps a regular InputStream, calling
// read() will skip some bytes as configured and will also
// return only data with configured length

class LimitedInputStream extends InputStream {

    private int skip;
    private int length;
    private InputStream is;

    LimitedInputStream(InputStream is, int skip, int length) {
        this.is = is;
        this.skip = skip;
        this.length = length;
    }

    @Override
    public int read() throws IOException {
        int r;
        while (skip > 0) {
            r = is.read();
            if (r < 0) {
                throw new IOException("stream ended before being able to skip all bytes");
            }
            skip--;
        }
        if (length == 0) {
            return -1;
        }
        r = is.read();
        if (r < 0) {
            throw new IOException("stream ended before being able to read all bytes");
        }
        length--;
        return r;
    }
}


