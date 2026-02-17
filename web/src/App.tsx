import { Toaster } from 'sonner';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Layout } from './components/Layout';
import { DashboardPage } from './pages/DashboardPage';
import { ContainersPage } from './pages/ContainersPage';
import { ImagesPage } from './pages/ImagesPage';
import { VolumesPage } from './pages/VolumesPage';

function App() {
  return (
    <BrowserRouter>
      <Toaster richColors position="top-center" />
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<DashboardPage />} />
          <Route path="dashboard" element={<DashboardPage />} />
          <Route path="containers" element={<ContainersPage />} />
          <Route path="images" element={<ImagesPage />} />
          <Route path="volumes" element={<VolumesPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
