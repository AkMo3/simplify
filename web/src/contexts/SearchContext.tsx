import { createContext, useContext, useState, type ReactNode } from 'react'

interface SearchContextValue {
  /** Current search query */
  query: string
  /** Update the search query */
  setQuery: (query: string) => void
  /** Clear the search query */
  clearQuery: () => void
}

const SearchContext = createContext<SearchContextValue | undefined>(undefined)

interface SearchProviderProps {
  children: ReactNode
}

export function SearchProvider({ children }: SearchProviderProps) {
  const [query, setQuery] = useState('')

  const clearQuery = () => setQuery('')

  return (
    <SearchContext.Provider value={{ query, setQuery, clearQuery }}>
      {children}
    </SearchContext.Provider>
  )
}

export function useSearch() {
  const context = useContext(SearchContext)
  if (context === undefined) {
    throw new Error('useSearch must be used within a SearchProvider')
  }
  return context
}
