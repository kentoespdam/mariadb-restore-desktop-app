import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './app';
import './style.css';

const host = document.getElementById('root');
if (!host) throw new Error('root element missing');
const root = createRoot(host);
root.render(
  <StrictMode>
    <App />
  </StrictMode>,
);
