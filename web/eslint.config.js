export default [
  {
    ignores: ['**/*.{ts,tsx,cts,mts,vue}', '**/*.d.ts']
  },
  {
    files: ['**/*.{js,jsx,cjs,mjs}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module'
    }
  }
]
