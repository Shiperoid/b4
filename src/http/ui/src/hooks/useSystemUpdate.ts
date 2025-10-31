import { useState, useCallback } from "react";

interface UpdateRequest {
  version?: string;
}

interface UpdateResponse {
  success: boolean;
  message: string;
  service_manager: string;
  update_command?: string;
}

export const useSystemUpdate = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const performUpdate = useCallback(
    async (version?: string): Promise<UpdateResponse | null> => {
      setLoading(true);
      setError(null);

      try {
        const response = await fetch("/api/system/update", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ version } as UpdateRequest),
        });

        const data = await response.json();

        if (!response.ok) {
          const errorMessage = data.message || "Failed to initiate update";
          setError(errorMessage);
          setLoading(false);
          return data;
        }

        return data;
      } catch (err) {
        console.log("Connection lost (expected during update)");
        return {
          success: true,
          message: "Update is in progress...",
          service_manager: "unknown",
        };
      }
    },
    []
  );

  const waitForReconnection = useCallback(
    async (maxAttempts: number = 60): Promise<boolean> => {
      let attempts = 0;

      while (attempts < maxAttempts) {
        try {
          await new Promise((resolve) => setTimeout(resolve, 3000));

          const response = await fetch("/api/version", {
            method: "GET",
            cache: "no-cache",
          });

          if (response.ok) {
            setLoading(false);
            return true;
          }
        } catch (err) {
          // Service not yet available
          attempts++;
        }
      }

      setLoading(false);
      setError("Update did not complete within expected time");
      return false;
    },
    []
  );

  return {
    performUpdate,
    waitForReconnection,
    loading,
    error,
  };
};
