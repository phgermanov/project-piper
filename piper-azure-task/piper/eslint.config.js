
// const neostandard = require('neostandard');

// ({
//   // options
// })

// eslint.config.js
// const eslint = require("@eslint/js");
// const tseslint = require("typescript-eslint");
// // import eslint from '@eslint/js';
// // import tseslint from 'typescript-eslint';

const globals = require("globals");

module.exports = require('neostandard')({
  ts: true,
  ignores: [
    "dist/**/*.js",
  ],

})

// module.exports = tseslint.config(
//   eslint.configs.recommended,
//   ...tseslint.configs.recommended,
//   {
//     // files: [
//     //   "**/*.ts",
//     //   "eslint.config.js",
//     // ],
//     // ignores: [
//     //   "dist/*",
//     //   // "dist/src/telemetry.js",
//     // ],
//     languageOptions: {
//         // ecmaVersion: 2022,
//         // sourceType: "module",
//         globals: {
//             // ...globals.browser,
//             //   es6: true,
//             ...globals.jest,
//             ...globals.node,
//         }
//     },
//     // extends:[
//     //   "standard-with-typescript",
//     //   "plugin:n/recommended"
//     // ],
//     // plugins:[
//     //   jest,
//     //   n
//     // ],
//     // parserOptions:{
//     //   project: "./tsconfig.json"
//     // },
//     rules: {
//       // "node/no-unsupported-features/es-syntax": [error, ignores: [modules]],
//       // "n/no-missing-import": ""off""
//       // FIXME fix eslint findings
//       // "eqeqeq": "off",
//       // "prefer-promise-reject-errors": "off",
//       // "new-cap": "off",
//       // "object-shorthand": "off",
//       // "n/no-unsupported-features/node-builtins": "off",
//       // "n/no-unpublished-import": "off",
//       // "n/no-unpublished-require": "off",
//       // "n/no-extraneous-require": "off",
//       // "n/no-unsupported-features/es-syntax": "off",
//       // "n/no-extraneous-import": "off",
//       // "@typescript-eslint/no-misused-promises": "off",
//       // "@typescript-eslint/no-unnecessary-type-assertion": "off",
//       // "@typescript-eslint/no-floating-promises": "off",
//       // "@typescript-eslint/no-this-alias": "off",
//       // "@typescript-eslint/no-non-null-assertion": "off",
//       // "@typescript-eslint/no-var-requires": "off",
//       // "@typescript-eslint/no-confusing-void-expression": "off",
//       // "@typescript-eslint/return-await": "off",
//       // "@typescript-eslint/ban-types": "off",
//       // "@typescript-eslint/prefer-readonly": "off",
//       // "@typescript-eslint/restrict-plus-operands": "off",
//       // "@typescript-eslint/ban-ts-comment": "off",
//       // "@typescript-eslint/prefer-nullish-coalescing": "off",
//       // "@typescript-eslint/prefer-ts-expect-error": "off",
//       // "@typescript-eslint/restrict-template-expressions": "off",
//       // "@typescript-eslint/explicit-function-return-type": "off",
//       // "@typescript-eslint/promise-function-async": ""off"",
//       // "@typescript-eslint/consistent-type-assertions": "off",
//       // "@typescript-eslint/strict-boolean-expressions": "off",
//       // "@typescript-eslint/await-thenable": "off",
//       "@typescript-eslint/no-unused-vars": "off",
//       "@typescript-eslint/no-explicit-any": "off",
//       "@typescript-eslint/no-require-imports": "off",
//       "@typescript-eslint/no-unsafe-function-type": "off",
//       "@typescript-eslint/no-unused-expressions": "off",
//       "@typescript-eslint/no-this-alias": "off",
//       "@typescript-eslint/ban-ts-comment": "off",
//       // "no-redeclare": "off",
//       "require-yield": "off",
//       "prefer-rest-params": "off",
//     }
//   }
// );
