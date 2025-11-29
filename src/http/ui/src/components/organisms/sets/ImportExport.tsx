import { useState, useEffect } from "react";
import { Alert, Box, Button, Stack } from "@mui/material";
import {
  Layers as LayersIcon,
  Refresh as RefreshIcon,
} from "@mui/icons-material";
import SettingSection from "@molecules/common/B4Section";
import B4TextField from "@atoms/common/B4TextField";
import { B4SetConfig } from "@models/Config";
import { button_secondary, button_primary } from "@design";

interface ImportExportSettingsProps {
  config: B4SetConfig;
  onImport: (importedConfig: B4SetConfig) => void;
}

export const ImportExportSettings: React.FC<ImportExportSettingsProps> = ({
  config,
  onImport,
}) => {
  const [jsonValue, setJsonValue] = useState("");
  const [originalJson, setOriginalJson] = useState("");
  const [validationError, setValidationError] = useState("");
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    const formatted = JSON.stringify(config, null, 0);
    setJsonValue(formatted);
    setOriginalJson(formatted);
    setValidationError("");
    setHasChanges(false);
  }, [config]);

  const handleJsonChange = (value: string) => {
    setJsonValue(value);
    setHasChanges(value !== originalJson);
    setValidationError("");
  };

  const handleValidate = () => {
    try {
      const parsed = JSON.parse(jsonValue) as B4SetConfig;

      // Validate required fields
      if (
        !parsed.name ||
        !parsed.tcp ||
        !parsed.udp ||
        !parsed.fragmentation ||
        !parsed.faking ||
        !parsed.targets
      ) {
        setValidationError(
          "Invalid set configuration: missing required fields"
        );
        return null;
      }

      setValidationError("");
      return parsed;
    } catch (error) {
      setValidationError(
        error instanceof Error ? error.message : "Invalid JSON format"
      );
      return null;
    }
  };

  const handleApply = () => {
    const validated = handleValidate();
    if (validated) {
      // Preserve the original ID
      validated.id = config.id;
      onImport(validated);
    }
  };

  const handleReset = () => {
    setJsonValue(originalJson);
    setHasChanges(false);
    setValidationError("");
  };

  return (
    <SettingSection title="JSON Configuration" icon={<LayersIcon />}>
      <Stack spacing={2}>
        <B4TextField
          label="Set Configuration JSON"
          value={jsonValue}
          onChange={(e) => handleJsonChange(e.target.value)}
          multiline
          rows={10}
          helperText="Edit directly or paste a configuration. Changes must be applied to take effect."
        />

        {validationError && <Alert severity="error">{validationError}</Alert>}

        <Box sx={{ display: "flex", gap: 2 }}>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={handleReset}
            disabled={!hasChanges}
            sx={{ ...button_secondary }}
          >
            Reset
          </Button>
          <Box sx={{ flex: 1 }} />
          <Button
            variant="outlined"
            onClick={handleValidate}
            sx={{ ...button_secondary }}
          >
            Validate
          </Button>
          <Button
            variant="contained"
            onClick={handleApply}
            disabled={!hasChanges}
            sx={{ ...button_primary }}
          >
            Apply Changes
          </Button>
        </Box>
      </Stack>
    </SettingSection>
  );
};
