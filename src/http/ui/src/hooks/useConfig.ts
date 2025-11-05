import { useState, useCallback, useEffect } from "react";
import B4Config from "../models/Config";

export function useLoadConfig() {
  const [config, setConfig] = useState<B4Config | null>(null);

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const response = await fetch("/api/config");
        if (!response.ok) throw new Error("Failed to load configuration");
        const data = await response.json();
        setConfig(data);
      } catch (error) {
        console.error("Error loading config:", error);
      }
    };

    fetchConfig();
  }, []);

  return { config };
}
