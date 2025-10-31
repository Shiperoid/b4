import React from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Alert,
  Typography,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  ListItemIcon,
  Radio,
  Stack,
  Box,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import DomainIcon from "@mui/icons-material/Language";
import { colors } from "../../../Theme";

interface DomainAddModalProps {
  open: boolean;
  domain: string;
  variants: string[];
  selected: string;
  onClose: () => void;
  onSelectVariant: (variant: string) => void;
  onAdd: () => void;
}

export const DomainAddModal: React.FC<DomainAddModalProps> = ({
  open,
  domain,
  variants,
  selected,
  onClose,
  onSelectVariant,
  onAdd,
}) => {
  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
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
              p: 1,
              borderRadius: 2,
              bgcolor: colors.accent.secondary,
              color: colors.secondary,
              display: "flex",
              alignItems: "center",
            }}
          >
            <DomainIcon />
          </Box>
          <Typography>Add Domain to Manual List</Typography>
        </Stack>
      </DialogTitle>
      <DialogContent sx={{ mt: 2 }}>
        <Alert severity="info" sx={{ mb: 2 }}>
          Select which domain pattern to add to the manual domains list. More
          specific patterns will only match exact subdomains, while broader
          patterns will match all subdomains.
        </Alert>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
          Original domain: <strong>{domain}</strong>
        </Typography>
        <List>
          {variants.map((variant, index) => (
            <ListItem key={variant} disablePadding>
              <ListItemButton
                onClick={() => onSelectVariant(variant)}
                selected={selected === variant}
                sx={{
                  borderRadius: 1,
                  mb: 0.5,
                  "&.Mui-selected": {
                    bgcolor: colors.accent.primary,
                    "&:hover": {
                      bgcolor: colors.accent.primaryHover,
                    },
                  },
                }}
              >
                <ListItemIcon>
                  <Radio
                    checked={selected === variant}
                    sx={{
                      color: colors.border.default,
                      "&.Mui-checked": {
                        color: colors.primary,
                      },
                    }}
                  />
                </ListItemIcon>
                <ListItemText
                  primary={variant}
                  secondary={
                    index === 0
                      ? "Most specific - exact match only"
                      : index === variants.length - 1
                      ? "Broadest - matches all subdomains"
                      : "Intermediate specificity"
                  }
                />
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      </DialogContent>
      <DialogActions
        sx={{ borderTop: `1px solid ${colors.border.light}`, p: 2 }}
      >
        <Button onClick={onClose} color="inherit">
          Cancel
        </Button>
        <Box sx={{ flex: 1 }} />
        <Button
          onClick={onAdd}
          variant="contained"
          startIcon={<AddIcon />}
          disabled={!selected}
          sx={{
            bgcolor: colors.primary,
            "&:hover": {
              bgcolor: colors.secondary,
            },
          }}
        >
          Add Domain
        </Button>
      </DialogActions>
    </Dialog>
  );
};
