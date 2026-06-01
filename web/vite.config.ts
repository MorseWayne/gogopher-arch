import { defineConfig } from 'vite'
import mdx from '@mdx-js/rollup'
import path from 'path'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'
import rehypeSlug from 'rehype-slug'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'


function figmaAssetResolver() {
  return {
    name: 'figma-asset-resolver',
    resolveId(id) {
      if (id.startsWith('figma:asset/')) {
        const filename = id.replace('figma:asset/', '')
        return path.resolve(__dirname, 'src/assets', filename)
      }
    },
  }
}

export default defineConfig({
  plugins: [
    figmaAssetResolver(),
    mdx({ remarkPlugins: [remarkGfm], rehypePlugins: [rehypeSlug, rehypeHighlight] }),
    // The React and Tailwind plugins are both required for Make, even if
    // Tailwind is not being actively used – do not remove them
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      // Alias @ to the src directory
      '@': path.resolve(__dirname, './src'),
    },
  },

  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },

  // File types to support raw imports. Never add .css, .tsx, or .ts files to this.
  assetsInclude: ['**/*.svg', '**/*.csv'],

  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return undefined
          }

          if (id.includes('/react-router/')) {
            return 'vendor-router'
          }
          if (id.includes('/@radix-ui/') || id.includes('/cmdk/') || id.includes('/vaul/')) {
            return 'vendor-ui'
          }
          if (id.includes('/@codemirror/') || id.includes('/@uiw/') || id.includes('/@lezer/')) {
            return 'vendor-editor'
          }
          if (id.includes('/lucide-react/')) {
            return 'vendor-icons'
          }
          if (id.includes('/recharts/') || id.includes('/d3-')) {
            return 'vendor-charts'
          }
          if (id.includes('/motion/')) {
            return 'vendor-motion'
          }
          if (id.includes('/@mdx-js/') || id.includes('/highlight.js/') || id.includes('/hast-util-') || id.includes('/rehype-') || id.includes('/remark-')) {
            return 'vendor-mdx'
          }

          return 'vendor'
        },
      },
    },
  },
})
