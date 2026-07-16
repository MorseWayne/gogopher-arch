import mdx from '@mdx-js/rollup'
import react from '@vitejs/plugin-react'
import rehypeHighlight from 'rehype-highlight'
import rehypeSlug from 'rehype-slug'
import remarkGfm from 'remark-gfm'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [
    mdx({ remarkPlugins: [remarkGfm], rehypePlugins: [rehypeSlug, rehypeHighlight] }),
    react(),
  ],
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.{ts,tsx}'],
    restoreMocks: true,
    setupFiles: ['./src/test/setup.ts'],
  },
})
