import { useEffect, useState } from "react";
import {
  Container,
  Box,
  Backdrop,
  CircularProgress,
  Stack,
  Typography,
  Snackbar,
  Alert,
  Button,
} from "@mui/material";
import { Save as SaveIcon, Refresh as RefreshIcon } from "@mui/icons-material";
import { SetsManager, SetWithStats } from "@organisms/sets/Manager";
import { B4Config, B4SetConfig } from "@models/Config";
import { colors } from "@design";

export default function Sets() {
  const [config, setConfig] = useState<
    (B4Config & { sets?: SetWithStats[] }) | null
  >(null);
  const [originalConfig, setOriginalConfig] = useState<B4Config | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [snackbar, setSnackbar] = useState<{
    open: boolean;
    message: string;
    severity: "success" | "error" | "info";
  }>({
    open: false,
    message: "",
    severity: "info",
  });

  const hasChanges =
    config && originalConfig
      ? JSON.stringify(config.sets) !== JSON.stringify(originalConfig.sets)
      : false;

  useEffect(() => {
    void loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      setLoading(true);
      const response = await fetch("/api/config");
      if (!response.ok) throw new Error("Failed to load");
      const data = (await response.json()) as B4Config & {
        sets?: SetWithStats[];
      };
      setConfig(data);
      setOriginalConfig(structuredClone(data));
    } catch {
      setSnackbar({
        open: true,
        message: "Failed to load configuration",
        severity: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const saveConfig = async () => {
    if (!config) return;
    try {
      setSaving(true);
      const response = await fetch("/api/config", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(config),
      });
      if (!response.ok) throw new Error(await response.text());
      setOriginalConfig(structuredClone(config));
      setSnackbar({
        open: true,
        message: "Sets saved successfully!",
        severity: "success",
      });
    } catch (error) {
      setSnackbar({
        open: true,
        message: error instanceof Error ? error.message : "Failed to save",
        severity: "error",
      });
    } finally {
      setSaving(false);
      await loadConfig();
    }
  };

  const handleChange = (
    field: string,
    value: boolean | string | number | B4SetConfig[]
  ) => {
    if (!config) return;
    if (field === "sets") {
      setConfig({ ...config, sets: value as SetWithStats[] });
    }
  };

  if (loading || !config) {
    return (
      <Backdrop open sx={{ zIndex: 9999 }}>
        <Stack alignItems="center" spacing={2}>
          <CircularProgress sx={{ color: colors.secondary }} />
          <Typography sx={{ color: colors.text.primary }}>
            Loading...
          </Typography>
        </Stack>
      </Backdrop>
    );
  }

  return (
    <Container
      maxWidth={false}
      sx={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
        py: 3,
      }}
    >
      <Box sx={{ display: "flex", justifyContent: "flex-end", mb: 2, gap: 1 }}>
        <Button
          size="small"
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={() => void loadConfig()}
          disabled={saving}
        >
          Reload
        </Button>
        <Button
          size="small"
          variant="contained"
          startIcon={saving ? <CircularProgress size={16} /> : <SaveIcon />}
          onClick={() => void saveConfig()}
          disabled={!hasChanges || saving}
          sx={{
            bgcolor: colors.secondary,
            "&:hover": { bgcolor: colors.primary },
            "&:disabled": { bgcolor: colors.accent.secondary },
          }}
        >
          {saving ? "Saving..." : "Save Changes"}
        </Button>
      </Box>

      <Box sx={{ flex: 1, overflow: "auto" }}>
        <SetsManager config={config} onChange={handleChange} />
      </Box>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
      >
        <Alert
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          severity={snackbar.severity}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Container>
  );
}
