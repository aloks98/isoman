import { BrowserRouter, Routes, Route } from 'react-router';
import { MainLayout } from './layouts/MainLayout';
import { Root } from './routes/Root';
import { IsosPage } from './routes/isos/IsosPage';
import { NotFound } from './routes/NotFound';
import './index.css';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Root />} />
        <Route element={<MainLayout />}>
          <Route path="/isos" element={<IsosPage />} />
        </Route>
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
  );
}
