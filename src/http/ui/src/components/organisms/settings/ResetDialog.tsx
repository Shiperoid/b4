// src/http/ui/src/components/organisms/settings/ResetDialog.tsx
import React, { useState } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Alert,
  Stack,
  Typography,
  Box,
  Divider,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  CircularProgress,
} from "@mui/material";
import {
  RestartAlt as ResetIcon,
  CheckCircle as CheckIcon,
  Error as ErrorIcon,
  Warning as WarningIcon,
  Shield as ShieldIcon,
} from "@mui/icons-material";
import { useConfigReset } from "../../../hooks/useConfig";
import { colors } from "../../../Theme";

interface ResetDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

type ResetState = "confirm" | "resetting" | "success" | "error";

export const ResetDialog: React.FC<ResetDialogProps> = ({
  open,
  onClose,
  onSuccess,
}) => {
  const [state, setState] = useState<ResetState>("confirm");
  const [message, setMessage] = useState("");
  const { resetConfig, loading } = useConfigReset();

  const handleReset = async () => {
    setState("resetting");
    setMessage("Resetting configuration...");

    const response = await resetConfig();

    if (response && response.success) {
      setState("success");
      setMessage("Configuration reset successfully!");
      setTimeout(() => {
        handleClose();
        onSuccess();
      }, 2000);
    } else {
      setState("error");
      setMessage("Failed to reset configuration");
    }
  };

  const handleClose = () => {
    if (state !== "resetting") {
      setState("confirm");
      setMessage("");
      onClose();
    }
  };

  const getDialogContent = () => {
    switch (state) {
      case "confirm":
        return (
          <>
            <DialogContent sx={{ mt: 2 }}>
              <Alert
                severity="warning"
                icon={<WarningIcon />}
                sx={{
                  bgcolor: colors.background.default,
                  border: `1px solid ${colors.quaternary}44`,
                  mb: 3,
                }}
              >
                <Typography variant="body2" sx={{ mb: 1 }}>
                  This will reset all configuration to default values except:
                </Typography>
              </Alert>

              <List dense>
                <ListItem>
                  <ListItemIcon>
                    <ShieldIcon sx={{ color: colors.secondary }} />
                  </ListItemIcon>
                  <ListItemText
                    primary="Domain Configuration"
                    secondary="All domain filters and geodata settings will be preserved"
                  />
                </ListItem>
                <ListItem>
                  <ListItemIcon>
                    <ShieldIcon sx={{ color: colors.secondary }} />
                  </ListItemIcon>
                  <ListItemText
                    primary="Testing Configuration"
                    secondary="Checker settings and test domains will be preserved"
                  />
                </ListItem>
              </List>

              <Alert
                severity="info"
                sx={{
                  mt: 2,
                  bgcolor: colors.background.default,
                  border: `1px solid ${colors.border.default}`,
                }}
              >
                <Typography variant="caption">
                  Network, DPI bypass, protocol, and logging settings will be
                  reset to defaults. You may need to restart B4 for some changes
                  to take effect.
                </Typography>
              </Alert>
            </DialogContent>

            <Divider sx={{ borderColor: colors.border.default }} />

            <DialogActions sx={{ p: 2 }}>
              <Button
                onClick={handleClose}
                sx={{
                  color: colors.text.secondary,
                  "&:hover": {
                    bgcolor: colors.accent.primaryHover,
                  },
                }}
              >
                Cancel
              </Button>
              <Box sx={{ flex: 1 }} />
              <Button
                onClick={handleReset}
                variant="contained"
                startIcon={<ResetIcon />}
                sx={{
                  bgcolor: colors.quaternary,
                  "&:hover": {
                    bgcolor: "#d32f2f",
                  },
                }}
              >
                Reset to Defaults
              </Button>
            </DialogActions>
          </>
        );

      case "resetting":
        return (
          <DialogContent sx={{ mt: 2 }}>
            <Stack spacing={3} alignItems="center" sx={{ py: 4 }}>
              <CircularProgress size={48} sx={{ color: colors.secondary }} />
              <Typography variant="h6" sx={{ color: colors.text.primary }}>
                {message}
              </Typography>
            </Stack>
          </DialogContent>
        );

      case "success":
        return (
          <DialogContent sx={{ mt: 2 }}>
            <Stack spacing={3} alignItems="center" sx={{ py: 4 }}>
              <CheckIcon
                sx={{
                  fontSize: 64,
                  color: colors.secondary,
                }}
              />
              <Typography variant="h6" sx={{ color: colors.text.primary }}>
                {message}
              </Typography>
            </Stack>
          </DialogContent>
        );

      case "error":
        return (
          <>
            <DialogContent sx={{ mt: 2 }}>
              <Stack spacing={3} alignItems="center" sx={{ py: 4 }}>
                <ErrorIcon sx={{ fontSize: 64, color: colors.quaternary }} />
                <Alert severity="error" sx={{ width: "100%" }}>
                  {message}
                </Alert>
              </Stack>
            </DialogContent>

            <Divider sx={{ borderColor: colors.border.default }} />

            <DialogActions sx={{ p: 2 }}>
              <Button onClick={handleClose} variant="contained">
                Close
              </Button>
            </DialogActions>
          </>
        );
    }
  };

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
      disableEscapeKeyDown={state === "resetting"}
      PaperProps={{
        sx: {
          bgcolor: colors.background.paper,
          border: `2px solid ${colors.border.default}`,
          borderRadius: 4,
        },
      }}
    >
      <DialogTitle
        sx={{
          bgcolor: colors.background.dark,
          color: colors.text.primary,
          borderBottom: `1px solid ${colors.border.default}`,
        }}
      >
        <Stack direction="row" alignItems="center" spacing={2}>
          <Box
            sx={{
              p: 1.5,
              borderRadius: 2,
              bgcolor: `${colors.quaternary}22`,
              color: colors.quaternary,
              display: "flex",
              alignItems: "center",
            }}
          >
            <ResetIcon />
          </Box>
          <Box>
            <Typography sx={{ mt: 1.5, lineHeight: 0 }}>
              Reset Configuration
            </Typography>
            <Typography variant="caption" sx={{ color: colors.text.secondary }}>
              Restore default settings
            </Typography>
          </Box>
        </Stack>
      </DialogTitle>
      {getDialogContent()}
    </Dialog>
  );
};
