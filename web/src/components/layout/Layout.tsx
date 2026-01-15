
import { Outlet, NavLink, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  Box,
  Activity,
  Settings,
  Search,
  X,
  Cuboid,
  Network,
} from 'lucide-react'
import { useSearch } from '@/contexts/SearchContext'
import { useRef, useEffect } from 'react'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/applications', icon: Box, label: 'Applications' },
  { to: '/pods', icon: Cuboid, label: 'Pods' },
  { to: '/networks', icon: Network, label: 'Networks' },
  { to: '/health', icon: Activity, label: 'Health' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

export function Layout() {
  const { query, setQuery, clearQuery } = useSearch()
  const searchInputRef = useRef<HTMLInputElement>(null)
  const location = useLocation()

  // Keyboard shortcut: Cmd/Ctrl + K to focus search
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        searchInputRef.current?.focus()
      }
      // Escape to clear and blur search
      if (e.key === 'Escape' && document.activeElement === searchInputRef.current) {
        clearQuery()
        searchInputRef.current?.blur()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [clearQuery])

  // Check if a nav item is active
  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/'
    return location.pathname.startsWith(path)
  }

  return (
    <div className="min-h-screen flex flex-col">
      {/* Top Navigation Bar - CONTRAST: Elevated surface with gradient */}
      <header className="navbar sticky top-0 z-50">
        <div className="flex h-14 items-center px-6 gap-8">
          {/* Logo - CONTRAST: White logo mark against dark bg */}
          <NavLink to="/" className="flex items-center gap-2.5 shrink-0 group">
            <div className="h-8 w-8 rounded-lg bg-[hsl(0_0%_95%)] flex items-center justify-center transition-transform duration-150 group-hover:scale-105">
              <span className="text-[hsl(0_0%_9%)] font-bold text-sm">S</span>
            </div>
            <span className="font-semibold text-[hsl(0_0%_90%)] tracking-tight">Simplify</span>
          </NavLink>

          {/* Navigation - CONTRAST: Clear active vs inactive states */}
          <nav className="flex items-center gap-1">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={isActive(item.to) ? 'nav-item-active' : 'nav-item'}
              >
                <item.icon className="h-4 w-4" />
                <span>{item.label}</span>
              </NavLink>
            ))}
          </nav>

          {/* Spacer */}
          <div className="flex-1" />

          {/* Search Bar - CONTRAST: Darker input against elevated navbar */}
          <div className="relative w-72">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            <input
              ref={searchInputRef}
              type="text"
              placeholder="Search applications..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="search-input w-full pl-9 pr-12"
            />
            {query ? (
              <button
                onClick={clearQuery}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
              >
                <X className="h-4 w-4" />
              </button>
            ) : (
              <kbd className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none hidden sm:inline-flex h-5 select-none items-center gap-1 rounded border border-[hsl(0_0%_25%)] bg-[hsl(0_0%_15%)] px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
                <span className="text-xs">⌘</span>K
              </kbd>
            )}
          </div>

          {/* Version Badge - CONTRAST: Subtle badge with border */}
          <span className="text-xs text-muted-foreground font-medium hidden md:block px-2 py-1 rounded bg-[hsl(0_0%_15%)] border border-[hsl(0_0%_20%)]">
            v0.1.0
          </span>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
