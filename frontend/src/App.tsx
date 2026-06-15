import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Spin } from 'antd';
import { Suspense, lazy } from 'react';
import { useAuthStore } from './store/auth';
import MainLayout from './components/Layout';
import RoleGuard from './components/RoleGuard';

// Lazy-loaded pages
const Login = lazy(() => import('./pages/Login'));
const Dashboard = lazy(() => import('./pages/Dashboard'));

const EquipList = lazy(() => import('./pages/equipment/List'));
const EquipDetail = lazy(() => import('./pages/equipment/Detail'));
const EquipCreate = lazy(() => import('./pages/equipment/Create'));
const EquipEdit = lazy(() => import('./pages/equipment/Edit'));

const UserList = lazy(() => import('./pages/users/List'));
const UserCreate = lazy(() => import('./pages/users/Create'));
const UserEdit = lazy(() => import('./pages/users/Edit'));
const ChangePassword = lazy(() => import('./pages/users/ChangePassword'));

const MyRecords = lazy(() => import('./pages/borrows/MyRecords'));
const PendingList = lazy(() => import('./pages/borrows/PendingList'));
const AllRecords = lazy(() => import('./pages/borrows/AllRecords'));
const BorrowApply = lazy(() => import('./pages/borrows/Apply'));

const NotFound = lazy(() => import('./pages/NotFound'));

function AuthGuard({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

function Loading() {
  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
      <Spin size="large" />
    </div>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <Suspense fallback={<Loading />}>
        <Routes>
          {/* Public */}
          <Route path="/login" element={<Login />} />

          {/* Protected routes wrapped in MainLayout */}
          <Route
            path="/"
            element={
              <AuthGuard>
                <MainLayout />
              </AuthGuard>
            }
          >
            <Route index element={<Dashboard />} />

            {/* Equipment */}
            <Route path="equipments" element={<EquipList />} />
            <Route path="equipments/:id" element={<EquipDetail />} />
            <Route
              path="equipments/new"
              element={
                <RoleGuard roles={['super_admin', 'lab_admin']}>
                  <EquipCreate />
                </RoleGuard>
              }
            />
            <Route
              path="equipments/:id/edit"
              element={
                <RoleGuard roles={['super_admin', 'lab_admin']}>
                  <EquipEdit />
                </RoleGuard>
              }
            />

            {/* Users */}
            <Route
              path="users"
              element={
                <RoleGuard roles={['super_admin', 'lab_admin']}>
                  <UserList />
                </RoleGuard>
              }
            />
            <Route
              path="users/new"
              element={
                <RoleGuard roles={['super_admin', 'lab_admin']}>
                  <UserCreate />
                </RoleGuard>
              }
            />
            <Route
              path="users/:id/edit"
              element={
                <RoleGuard roles={['super_admin', 'lab_admin']}>
                  <UserEdit />
                </RoleGuard>
              }
            />
            <Route path="users/:id/password" element={<ChangePassword />} />

            {/* Borrows */}
            <Route path="borrows/my" element={<MyRecords />} />
            <Route path="borrows/apply" element={<BorrowApply />} />
            <Route
              path="borrows/pending"
              element={
                <RoleGuard roles={['super_admin', 'lab_admin']}>
                  <PendingList />
                </RoleGuard>
              }
            />
            <Route
              path="borrows/all"
              element={
                <RoleGuard roles={['super_admin', 'lab_admin']}>
                  <AllRecords />
                </RoleGuard>
              }
            />
          </Route>

          {/* 404 */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
