import '@testing-library/jest-dom';
import { assert, describe, it } from 'vitest';
import { extractDomainFromHostname, isValidRedirectURLDomain } from '../../../views/pages/SignIn';

//-----------------------------------------------
// test
//-----------------------------------------------
describe('extractDomainFromHostname', () => {
  it('returns domain from hostname', () => {
    assert.equal(extractDomainFromHostname('dashboard.cosmo.github.io'), 'cosmo.github.io');
    assert.equal(extractDomainFromHostname('cosmo-dashboard.github.com'), 'github.com');
    assert.equal(extractDomainFromHostname('main-ws1-tom-k3d-code-server.example.cosmo.github.io'), 'example.cosmo.github.io');
    assert.equal(extractDomainFromHostname('localhost'), 'localhost');
  });
});

describe('isValidRedirectURLDomain', () => {
  beforeEach(() => {
    mockWindowHostname('localhost');
  });
  describe('when redirect url has the same domain as current url', () => {
    it('returns true', () => {
      mockWindowHostname('dashboard.cosmo.github.io');
      assert.equal(isValidRedirectURLDomain('https://tom-workspace1.cosmo.github.io/api/foo/bar'), true);
    });
  });

  describe('when redirect url with port has the same domain as current url', () => {
    it('returns true', () => {
      mockWindowHostname('cosmo-dashboard.github.com');
      assert.equal(isValidRedirectURLDomain('wss://main-workspace-tom.github.com:3000/api/foo/bar'), true);
    });
  });

  describe('when redirect url which subdomain is long has the same domain as current url', () => {
    it('returns true', () => {
      mockWindowHostname('dashboard.example.cosmo.github.io');
      assert.equal(isValidRedirectURLDomain('http://main-ws1-tom-k3d-code-server.example.cosmo.github.io/api/foo/bar'), true);
    });
  });

  describe('when current url and redirect url are localhost', () => {
    it('returns true', () => {
      assert.equal(isValidRedirectURLDomain('http://localhost:5000/api/foo/bar'), true);
    });
  });

  describe('when redirect url does NOT have the same domain as current url', () => {
    it('returns false', () => {
      mockWindowHostname('dashboard.cosmo.github.io');
      assert.equal(isValidRedirectURLDomain('https://tom-workspace1.c0smo.github.io/api/foo/bar'), false);
    });
  });
});

const mockWindowHostname = (hostname: string) => {
  global.window = Object.create(window);
  Object.defineProperty(window, 'location', {
    value: {
      hostname: hostname,
    },
  });
}