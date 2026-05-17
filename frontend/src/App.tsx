import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Header from './components/Header';
import CardConfirmPage from './pages/CardConfirmPage';
import CardRegisterPage from './pages/CardRegisterPage';
import PurchasePage from './pages/PurchasePage';
import PurchaseSuccessPage from './pages/PurchaseSuccessPage';
import PurchaseFailurePage from './pages/PurchaseFailurePage';
import PaymentListPage from './pages/PaymentListPage';

export default function App() {
  return (
    <BrowserRouter>
      <Header />
      <main>
        <Routes>
          <Route path="/card" element={<CardConfirmPage />} />
          <Route path="/card/register" element={<CardRegisterPage />} />
          <Route path="/purchase" element={<PurchasePage />} />
          <Route path="/purchase/success" element={<PurchaseSuccessPage />} />
          <Route path="/purchase/failure" element={<PurchaseFailurePage />} />
          <Route path="/payments" element={<PaymentListPage />} />
          <Route path="*" element={<Navigate to="/card" replace />} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}
