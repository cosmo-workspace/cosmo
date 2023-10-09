import { describe, expect, it } from "vitest";
import { base64url } from '../../components/Base64';

//-----------------------------------------------
// test
//-----------------------------------------------

function toArrayBuffer(buffer) {
  const arrayBuffer = new ArrayBuffer(buffer.length);
  const view = new Uint8Array(arrayBuffer);
  for (let i = 0; i < buffer.length; ++i) {
    view[i] = buffer[i];
  }
  return arrayBuffer;
}

function toBuffer(arrayBuffer) {
  const buffer = Buffer.alloc(arrayBuffer.byteLength);
  const view = new Uint8Array(arrayBuffer);
  for (let i = 0; i < buffer.length; ++i) {
    buffer[i] = view[i];
  }
  return buffer;
}

describe('base64url', () => {
  describe('encode', () => {
    it('✅ ok', async () => {
      const tests = [
        {
          name: "1",
          input: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/",
        },
        {
          name: "1",
          input: "a",
        },
        {
          name: "2",
          input: "aa",
        },
        {
          name: "3",
          input: "aaa",
        },
        {
          name: "4",
          input: "aaaa",
        },
        {
          name: "5",
          input: "aaaaa",
        },
      ];
      for (const t of tests) {
        const raw = Buffer.from(t.input, 'utf8');
        const got = base64url.encode(raw);

        const want = raw.toString('base64url');
        expect(got).toEqual(want);
      }
    });
  });

  describe('decode', () => {
    it('✅ ok', async () => {
      const base64str = 'QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVphYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ejAxMjM0NTY3ODkrLw==';
      const got = base64url.decode(base64str);
      const want = Buffer.from(base64str, 'base64url');
      expect(got).toEqual(toArrayBuffer(want));
    });

    it('✅ ok: no padding', async () => {
      const base64strWithPadding = 'QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVphYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ejAxMjM0NTY3ODkrLw==';
      const want = Buffer.from(base64strWithPadding, 'base64url');

      const base64strNoPadding = 'QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVphYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ejAxMjM0NTY3ODkrLw';
      const got = base64url.decode(base64strNoPadding);
      expect(got).toEqual(toArrayBuffer(want));
    });
  });

});
