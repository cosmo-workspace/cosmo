const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
  app.use('/api', createProxyMiddleware({ 
    target: 'http://localhost:8443',
    // pathRewrite: {'^/api': ''},
    changeOrigin: true 
  }));
};