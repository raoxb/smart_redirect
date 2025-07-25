import React, { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider, theme } from 'antd'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { useAuthStore } from '@/store/authStore'

// Layouts
import MainLayout from '@/components/Layout/MainLayout'

// Pages
import Login from '@/pages/Login'
import Dashboard from '@/pages/Dashboard'
import LinksPage from '@/pages/Links'
import LinkDetail from '@/pages/Links/LinkDetail'
import Statistics from '@/pages/Statistics'
import Templates from '@/pages/Templates'
import Users from '@/pages/Users'
import Monitoring from '@/pages/Monitoring'
import AccessLogs from '@/pages/AccessLogs'

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
})

// Protected Route Component
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore()
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
}

function App() {
  const { checkAuth } = useAuthStore()
  const { defaultAlgorithm, darkAlgorithm } = theme
  const [isDarkMode] = React.useState(false)

  useEffect(() => {
    checkAuth()
  }, [checkAuth])

  return (
    <ConfigProvider
      theme={{
        algorithm: isDarkMode ? darkAlgorithm : defaultAlgorithm,
        token: {
          colorPrimary: '#1890ff',
        },
      }}
    >
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <MainLayout />
                </ProtectedRoute>
              }
            >
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="links" element={<LinksPage />} />
              <Route path="links/:linkId" element={<LinkDetail />} />
              
              {/* Add more routes here */}
              <Route path="statistics" element={<Statistics />} />
              <Route path="templates" element={<Templates />} />
              <Route path="users" element={<Users />} />
              <Route path="monitoring" element={<Monitoring />} />
              <Route path="access-logs" element={<AccessLogs />} />
              <Route path="profile" element={<div>Profile Page (Coming Soon)</div>} />
              <Route path="settings" element={<div>Settings Page (Coming Soon)</div>} />
            </Route>
            
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </BrowserRouter>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </ConfigProvider>
  )
}

export default App