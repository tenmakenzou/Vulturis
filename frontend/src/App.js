import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Nav from './Compoments/Nav/Nav';
import Login from './Pages/Login/LoginRegister.jsx';
import Checkout from './Pages/Checkout/Checkout.js';
import Contact from './Pages/Contact/Contact.js';
import Home from './Pages/Home/Home.jsx';
import Cart from './Pages/Cart/Cart.js';
import { AuthProvider } from './AuthContext';
import { CartProvider } from './CartContext';


function App() {
  return (
    <AuthProvider>
      <CartProvider> 
        <Router>
          <Nav />
          <main>
            <Routes>
              <Route path="/" element={<Home />} />
              <Route path="/contact" element={<Contact />} />
              <Route path="/login" element={<Login />} />
              <Route path="/cart" element={<Cart />} />
              <Route path="/checkout" element={<Checkout />} />
            </Routes>
          </main>
        </Router>
      </CartProvider>
    </AuthProvider>
  );
}


export default App;

