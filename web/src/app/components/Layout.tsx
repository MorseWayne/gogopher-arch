import { Outlet, Link } from "react-router";
import { Terminal, Github, Menu, X } from "lucide-react";
import { useState } from "react";
import { motion, AnimatePresence } from "motion/react";

export function Layout() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  return (
    <div className="min-h-screen bg-neutral-950 text-neutral-50 font-sans selection:bg-[#00ADD8] selection:text-white flex flex-col">
      <nav className="sticky top-0 z-50 border-b border-neutral-800 bg-neutral-950/80 backdrop-blur-md">
        <div className="container mx-auto px-6 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 text-xl font-bold text-[#00ADD8]">
            <Terminal className="w-6 h-6" />
            <span>GoGopher Arch</span>
          </Link>
          
          <div className="hidden md:flex items-center gap-8">
            <Link to="/" className="text-sm font-medium text-neutral-300 hover:text-white transition-colors">首页</Link>
            <a href="#features" className="text-sm font-medium text-neutral-300 hover:text-white transition-colors">核心特性</a>
            <a href="#roadmap" className="text-sm font-medium text-neutral-300 hover:text-white transition-colors">路线图</a>
            
            {/* User Profile Mock */}
            <div className="h-6 w-px bg-neutral-800 mx-2"></div>
            
            <Link to="/dashboard" className="flex items-center gap-3 group">
              <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-[#00ADD8] to-purple-500 p-0.5">
                <div className="w-full h-full bg-neutral-900 rounded-full border border-transparent group-hover:border-white/20 transition-all flex items-center justify-center text-xs font-bold text-white overflow-hidden">
                  <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=Gopher&backgroundColor=171717" alt="avatar" />
                </div>
              </div>
              <div className="flex flex-col">
                <span className="text-xs font-bold text-white leading-none mb-1">Gopher 实习生</span>
                <div className="flex items-center gap-1">
                  <div className="w-16 h-1 bg-neutral-800 rounded-full overflow-hidden">
                    <div className="h-full bg-[#00ADD8] w-[15%]"></div>
                  </div>
                  <span className="text-[10px] text-[#00ADD8] font-mono leading-none">Lv.1</span>
                </div>
              </div>
            </Link>
            <a 
              href="https://github.com/MorseWayne/gogopher-arch" 
              target="_blank" 
              rel="noreferrer"
              className="flex items-center gap-2 bg-white text-neutral-950 px-4 py-2 rounded-lg text-sm font-bold hover:bg-neutral-200 transition-all hover:scale-105"
            >
              <Github className="w-4 h-4" />
              GitHub
            </a>
          </div>

          <button 
            className="md:hidden text-neutral-300 hover:text-white"
            onClick={() => setIsMenuOpen(!isMenuOpen)}
          >
            {isMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>
      </nav>

      <AnimatePresence>
        {isMenuOpen && (
          <motion.div 
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            className="md:hidden bg-neutral-900 border-b border-neutral-800"
          >
            <div className="flex flex-col p-4 gap-4">
              <Link to="/" className="text-neutral-300 hover:text-white font-medium px-4 py-2 rounded-lg hover:bg-neutral-800 transition-colors" onClick={() => setIsMenuOpen(false)}>首页</Link>
              <Link to="/dashboard" className="text-neutral-300 hover:text-white font-medium px-4 py-2 rounded-lg hover:bg-neutral-800 transition-colors" onClick={() => setIsMenuOpen(false)}>进入沙盒</Link>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <main className="flex-1 flex flex-col">
        <Outlet />
      </main>

      <footer className="border-t border-neutral-900 bg-neutral-950 py-12">
        <div className="container mx-auto px-6 flex flex-col md:flex-row items-center justify-between text-neutral-500 text-sm">
          <div className="flex items-center gap-2 mb-4 md:mb-0">
            <Terminal className="w-5 h-5" />
            <span>© 2026 GoGopher Arch. MIT License.</span>
          </div>
          <div className="flex gap-6">
            <a href="#" className="hover:text-neutral-300 transition-colors">关于</a>
            <a href="#" className="hover:text-neutral-300 transition-colors">隐私政策</a>
            <a href="#" className="hover:text-neutral-300 transition-colors">条款</a>
          </div>
        </div>
      </footer>
    </div>
  );
}
