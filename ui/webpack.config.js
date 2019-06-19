/*
  - GIT_SHA, VERSION, etc. constants
  - static assets
*/

const path = require('path')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')

const mode =
  process.env.NODE_ENV === 'production' ? 'production' : 'development'

module.exports = {
  entry: path.resolve(__dirname, 'src', 'index.tsx'),
  devtool: 'inline-source-map',
  mode,
  output: {
    filename:
      mode === 'production'
        ? '[name].[chunkhash].js'
        : '[name].[chunkhash].dev.js',
    path: path.resolve(__dirname, 'build'),
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        // TODO: fork loader
        // TODO: Babel
        use: ['ts-loader'],
        exclude: /(node_modules)|(\.test\.tsx?$)/,
      },
      {
        test: /\.s?css/,
        // TODO: Minicssextract plugin prod
        use: [
          mode === 'production' ? MiniCssExtractPlugin.loader : 'style-loader',
          'style-loader',
          'css-loader',
          {
            loader: 'sass-loader',
            options: {
              implementation: require('sass'),
            },
          },
        ],
      },
      {
        test: /\.(ico|png|cur|jpg|ttf|eot|svg|woff(2)?)(\?[a-z0-9]+)?$/,
        loader: 'file-loader',
      },
    ],
  },
  resolve: {
    extensions: ['.ts', '.tsx', '.js'],
    alias: {
      // Allow global import paths
      src: path.resolve(__dirname, 'src'),
    },
  },
  plugins: [new MiniCssExtractPlugin()],
}
