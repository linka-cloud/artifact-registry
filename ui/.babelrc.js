const makePlugins = () => {
  if (process.env.NODE_ENV === 'production' && !process.env.DEBUG_BUILD) {
    console.log('production build: enable transform-remove-console')
    return [
      ['transform-remove-console', { exclude: ['error'] }],
      ['babel-plugin-direct-import', { modules: ['@mui/material', '@mui/icons-material'] }],
    ]
  }
  console.log('debug build: disable transform-remove-console')
  return [['babel-plugin-direct-import', { modules: ['@mui/material', '@mui/icons-material'] }]]
}

module.exports = { plugins: makePlugins() }
