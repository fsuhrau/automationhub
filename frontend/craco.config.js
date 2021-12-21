// craco.config.js
const CracoEsbuildPlugin = require('craco-esbuild');

module.exports = {
    webpack: {
        configure: {
            target: 'web'
        },
    },
    plugins: [
        {
            plugin: CracoEsbuildPlugin,
            options: {
                // includePaths: ['/external/dir/with/components'], // Optional. If you want to include components which are not in src folder
                enableSvgr: true, // Optional.
                svgrOptions: {
                    // Optional. is enableSvgr set to true, used as options for svgr
                    icon: true,
                },
                esbuildLoaderOptions: {
                    // Optional. Defaults to auto-detect loader.
                    loader: 'jsx', // Set the value to 'tsx' if you use typescript
                    target: 'es2015',
                },
                esbuildMinimizerOptions: {
                    // Optional. Defaults to:
                    target: 'es2015',
                    css: true, // if true, OptimizeCssAssetsWebpackPlugin will also be replaced by esbuild.
                },
                skipEsbuildJest: false, // Optional. Set to true if you want to use babel for jest tests,
                esbuildJestOptions: {
                    loaders: {
                        '.ts': 'ts',
                        '.tsx': 'tsx',
                    },
                },
            },
        },
    ],
};