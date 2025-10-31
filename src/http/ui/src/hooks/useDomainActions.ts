import { useState, useCallback } from "react";

interface DomainModalState {
  open: boolean;
  domain: string;
  variants: string[];
  selected: string;
}

interface SnackbarState {
  open: boolean;
  message: string;
  severity: "success" | "error";
}

export function useDomainActions() {
  const [modalState, setModalState] = useState<DomainModalState>({
    open: false,
    domain: "",
    variants: [],
    selected: "",
  });

  const [snackbar, setSnackbar] = useState<SnackbarState>({
    open: false,
    message: "",
    severity: "success",
  });

  const openModal = useCallback((domain: string, variants: string[]) => {
    setModalState({
      open: true,
      domain,
      variants,
      selected: variants[0] || domain,
    });
  }, []);

  const closeModal = useCallback(() => {
    setModalState({
      open: false,
      domain: "",
      variants: [],
      selected: "",
    });
  }, []);

  const selectVariant = useCallback((variant: string) => {
    setModalState((prev) => ({ ...prev, selected: variant }));
  }, []);

  const addDomain = useCallback(async () => {
    if (!modalState.selected) return;

    try {
      const response = await fetch("/api/geosite/domain", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          domain: modalState.selected,
        }),
      });

      if (response.ok) {
        setSnackbar({
          open: true,
          message: `Successfully added "${modalState.selected}" to manual domains`,
          severity: "success",
        });
        closeModal();
      } else {
        const error = await response.json();
        setSnackbar({
          open: true,
          message: `Failed to add domain: ${error.message}`,
          severity: "error",
        });
      }
    } catch (error) {
      setSnackbar({
        open: true,
        message: `Error adding domain: ${error}`,
        severity: "error",
      });
    }
  }, [modalState.selected, closeModal]);

  const closeSnackbar = useCallback(() => {
    setSnackbar((prev) => ({ ...prev, open: false }));
  }, []);

  return {
    modalState,
    snackbar,
    openModal,
    closeModal,
    selectVariant,
    addDomain,
    closeSnackbar,
  };
}
