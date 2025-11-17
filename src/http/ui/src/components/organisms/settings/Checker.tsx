import { Grid } from "@mui/material";
import { Science as TestIcon } from "@mui/icons-material";
import SettingSection from "@molecules/common/B4Section";
import B4Slider from "@atoms/common/B4Slider";
import { B4Config } from "@models/Config";

interface CheckerSettingsProps {
  config: B4Config;
  onChange: (
    field: string,
    value: string | boolean | number | string[]
  ) => void;
}

export const CheckerSettings: React.FC<CheckerSettingsProps> = ({
  config,
  onChange,
}) => {
  return (
    <SettingSection
      title="Testing Configuration"
      description="Configure testing behavior and output"
      icon={<TestIcon />}
    >
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 6 }}>
          <B4Slider
            label="Max Concurrent Tests"
            value={config.system.checker.max_concurrent}
            onChange={(value) =>
              onChange("system.checker.max_concurrent", value)
            }
            min={1}
            max={20}
            step={1}
            helperText="Maximum number of concurrent tests"
          />
        </Grid>
        <Grid size={{ xs: 12, lg: 6 }}>
          <B4Slider
            label="Test Timeout"
            value={config.system.checker.timeout}
            onChange={(value) => onChange("system.checker.timeout", value)}
            min={1}
            max={120}
            step={1}
            valueSuffix=" sec"
            helperText="Domain request timeout"
          />
        </Grid>
      </Grid>
    </SettingSection>
  );
};
