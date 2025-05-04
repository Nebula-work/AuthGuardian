import React from 'react';
import { Route, Routes, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';

// Pages
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Users from './pages/Users';
import Organizations from './pages/Organizations';
import Roles from './pages/Roles';
import Documentation from './pages/Documentation';

// Components
import Navbar from './components/Navbar';
import Sidebar from './components/Sidebar';
import { DashboardLayout } from './components/Layout/DashboardLayout';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

// Protected route component
const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
    const { isAuthenticated, loading } = useAuth();

    if (loading) {
        return <div className="flex items-center justify-center h-screen">Loading...</div>;
    }

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }


    return <DashboardLayout>{children}</DashboardLayout>;

};


function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Dashboard />
          </ProtectedRoute>
        }
      />
      <Route
        path="/users"
        element={
          <ProtectedRoute>
            <Users />
          </ProtectedRoute>
        }
      />
      <Route
        path="/organizations"
        element={
          <ProtectedRoute>
            <Organizations />
          </ProtectedRoute>
        }
      />
      <Route
        path="/roles"
        element={
          <ProtectedRoute>
            <Roles />
          </ProtectedRoute>
        }
      />
      {/*<Route*/}
      {/*  path="/documentation"*/}
      {/*  element={*/}
      {/*    <ProtectedRoute>*/}
      {/*      <Documentation />*/}
      {/*    </ProtectedRoute>*/}
      {/*  }*/}
      {/*/>*/}
      {/*<Route*/}
      {/*  path="/documentation/:docId"*/}
      {/*  element={*/}
      {/*    <ProtectedRoute>*/}
      {/*      <Documentation />*/}
      {/*    </ProtectedRoute>*/}
      {/*  }*/}
      {/*/>*/}
      <Route path="*" element={<Navigate to="/" replace />} />
      <Route path="/docs" element={<Documentation/>} />
      <Route path="/docs/:docId" element={<Documentation />} />

    </Routes>
  );
}

export default App;