const { defineConfig } = require('@vue/cli-service')


console.log(process.env.VUE_APP_API_HOST)

module.exports = defineConfig({
  transpileDependencies: true,
  publicPath: process.env.NODE_ENV === 'production' ? '/static' : ''
})
