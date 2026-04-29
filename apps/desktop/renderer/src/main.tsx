import "@bifrost/design-tokens/app.css";

import React from "react";
import ReactDOM from "react-dom/client";

import { AppProviders } from "./app/providers";

const rootElement = document.getElementById("app");

if (!rootElement) {
  throw new Error("Bifrost desktop root element was not found.");
}

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <AppProviders />
  </React.StrictMode>,
);
