import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { SearchProvider } from '@/contexts'
import { Layout } from '@/components/layout/Layout'
import { Dashboard } from '@/pages/Dashboard'
import { Applications } from '@/pages/Applications'
import { ApplicationDetail } from '@/pages/ApplicationDetail'
import { Health } from '@/pages/Health'
import { Settings } from '@/pages/Settings'

function App() {
  return (
    <SearchProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Dashboard />} />
            <Route path="applications" element={<Applications />} />
            <Route path="applications/:id" element={<ApplicationDetail />} />
            <Route path="health" element={<Health />} />
            <Route path="settings" element={<Settings />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </SearchProvider>
  )
}

export default App
