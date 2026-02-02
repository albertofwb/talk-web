import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login'
import Talk from './pages/Talk'
import Admin from './pages/Admin'
import { isAuthenticated, isAdmin } from './utils/auth'

function PrivateRoute({ children }: { children: JSX.Element }) {
  return isAuthenticated() ? children : <Navigate to="/login" />
}

function AdminRoute({ children }: { children: JSX.Element }) {
  return isAuthenticated() && isAdmin() ? children : <Navigate to="/talk" />
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/talk"
          element={
            <PrivateRoute>
              <Talk />
            </PrivateRoute>
          }
        />
        <Route
          path="/admin"
          element={
            <AdminRoute>
              <Admin />
            </AdminRoute>
          }
        />
        <Route path="/" element={<Navigate to="/talk" />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
