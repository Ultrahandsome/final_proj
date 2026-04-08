// /**
//  * @name 代理的配置
//  * @see 在生产环境 代理是无法生效的，所以这里没有生产环境的配置
//  * -------------------------------
//  * The agent cannot take effect in the production environment
//  * so there is no configuration of the production environment
//  * For details, please see
//  * https://pro.ant.design/docs/deploy
//  *
//  * @doc https://umijs.org/docs/guides/proxy
//  */
// export default {
//   // 如果需要自定义本地开发服务器  请取消注释按需调整
//   dev: {
//     // localhost:8000/api/** -> https://preview.pro.ant.design/api/**
//     '/api/': {
//       // 要代理的地址
//       target: 'http://localhost:38080',
//       // 配置了这个可以从 http 代理到 https
//       // 依赖 origin 的功能可能需要这个，比如 cookie
//       changeOrigin: true,
//     },
//   },
//   /**
//    * @name 详细的代理配置
//    * @doc https://github.com/chimurai/http-proxy-middleware
//    */
//   test: {
//     // localhost:8000/api/** -> https://preview.pro.ant.design/api/**
//     '/api/': {
//       target: 'http://localhost:38080',
//       changeOrigin: true,
//       pathRewrite: { '^': '' },
//     },
//   },
//   pre: {
//     '/api/': {
//       target: 'http://localhost:38080',
//       changeOrigin: true,
//       pathRewrite: { '^': '' },
//     },
//   },
// };


/**
 * @name 代理的配置
 * @description 本地开发环境的 API 请求代理设置
 * @see https://umijs.org/docs/guides/proxy
 *
 * @note 注意：代理在生产环境中不会生效，仅开发调试时可用
 */

export default {
  dev: {
    '/api': {
      target: 'http://localhost:38080', // 你的后端服务地址
      changeOrigin: true,
      pathRewrite: { '^/api': '' }, // 将 /api 前缀去除，转发到后端真实路径
    },
  },
  test: {
    '/api': {
      target: 'http://localhost:38080',
      changeOrigin: true,
      pathRewrite: { '^/api': '' },
    },
  },
  pre: {
    '/api': {
      target: 'http://localhost:38080',
      changeOrigin: true,
      pathRewrite: { '^/api': '' },
    },
  },
};
