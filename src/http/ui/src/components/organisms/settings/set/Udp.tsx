import React from "react";
import { Grid } from "@mui/material";
import { Dns as DnsIcon } from "@mui/icons-material";
import SettingSection from "@molecules/common/B4Section";
import SettingSelect from "@atoms/common/B4Select";
import SettingTextField from "@atoms/common/B4TextField";
import B4Slider from "@atoms/common/B4Slider";
import { B4SetConfig } from "@models/Config";
import B4Switch from "@/components/atoms/common/B4Switch";

interface UdpSettingsProps {
  config: B4SetConfig;
  onChange: (field: string, value: string | boolean | number) => void;
}

const UDP_MODES = [
  { value: "drop", label: "Drop" },
  { value: "fake", label: "Fake" },
];

const UDP_FAKING_STRATEGIES = [
  { value: "none", label: "None" },
  { value: "ttl", label: "TTL" },
  { value: "checksum", label: "Checksum" },
];

const UDP_QUIC_FILTERS = [
  { value: "disabled", label: "Disabled" },
  { value: "all", label: "All" },
  { value: "parse", label: "Parse" },
];

export const UdpSettings: React.FC<UdpSettingsProps> = ({
  config,
  onChange,
}) => {
  return (
    <SettingSection
      title="UDP Configuration"
      description="Configure UDP packet handling and QUIC filtering"
      icon={<DnsIcon />}
    >
      <Grid container spacing={3}>
        <Grid size={{ xs: 12, md: 6 }}>
          <SettingSelect
            label="UDP Mode"
            value={config.udp.mode}
            options={UDP_MODES}
            onChange={(e) => onChange("udp.mode", e.target.value as string)}
            helperText="UDP packet handling strategy"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <SettingSelect
            label="QUIC Filter"
            value={config.udp.filter_quic}
            options={UDP_QUIC_FILTERS}
            onChange={(e) =>
              onChange("udp.filter_quic", e.target.value as string)
            }
            helperText="QUIC traffic filtering mode"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <SettingSelect
            label="Faking Strategy"
            value={config.udp.faking_strategy}
            options={UDP_FAKING_STRATEGIES}
            onChange={(e) =>
              onChange("udp.faking_strategy", e.target.value as string)
            }
            helperText="Strategy for fake UDP packets"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <B4Switch
            label="Ignore STUN Packets"
            checked={config.udp.filter_stun}
            onChange={(checked) => onChange("udp.filter_stun", checked)}
            description="When enabled, STUN packets will be ignored"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <B4Slider
            label="Fake Packet Size"
            value={config.udp.fake_len}
            onChange={(value) => onChange("udp.fake_len", value)}
            min={32}
            max={1500}
            step={8}
            valueSuffix=" bytes"
            helperText="Size of fake UDP packets"
          />
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <B4Slider
            label="Fake Sequence Length"
            value={config.udp.fake_seq_length}
            onChange={(value) => onChange("udp.fake_seq_length", value)}
            min={1}
            max={20}
            step={1}
            helperText="Number of fake packets to send"
          />
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <SettingTextField
            label="Destination Port Filter"
            value={config.udp.dport_filter}
            onChange={(e) => onChange("udp.dport_filter", e.target.value)}
            helperText="Destination port filter, e.g., 80,443,1000-2000 (from 1 to 65535)"
          />
        </Grid>
      </Grid>
    </SettingSection>
  );
};
