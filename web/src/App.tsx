import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { SearchProvider } from '@/contexts'
import { Layout } from '@/components/layout/Layout'
import { Dashboard } from '@/pages/Dashboard'
import { Applications } from '@/pages/Applications'
import { ApplicationDetail } from '@/pages/ApplicationDetail'
import { Pods } from '@/pages/Pods'
import PodDetail from '@/pages/PodDetail'
import { Health } from '@/pages/Health'
import { Settings } from '@/pages/Settings'
import Networks from '@/pages/Networks'

function App() {
  return (
    <SearchProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Dashboard />} />
            <Route path="applications" element={<Applications />} />
            <Route path="applications/:id" element={<ApplicationDetail />} />
            <Route path="pods" element={<Pods />} />
            <Route path="pods/:id" element={<PodDetail />} />
            <Route path="networks" element={<Networks />} />
            <Route path="health" element={<Health />} />
            <Route path="settings" element={<Settings />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </SearchProvider>
  )
}

export default App
