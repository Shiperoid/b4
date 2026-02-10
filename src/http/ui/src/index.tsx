import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router";
import App from "./App";
import { WebSocketProvider } from "./context/B4WsProvider";

const root = createRoot(document.getElementById("root")!);
root.render(
  <BrowserRouter>
    <WebSocketProvider>
      <App />
    </WebSocketProvider>
  </BrowserRouter>,
);
