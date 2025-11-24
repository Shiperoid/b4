import {
  Grid,
  FormControlLabel,
  Switch,
  Typography,
  Divider,
  Chip,
} from "@mui/material";
import { Dns as DnsIcon } from "@mui/icons-material";
import SettingSection from "@molecules/common/B4Section";
import B4Slider from "@atoms/common/B4Slider";
import { B4SetConfig, WindowMode, DesyncMode } from "@models/Config";
import SettingSelect from "@atoms/common/B4Select";

interface TcpSettingsProps {
  config: B4SetConfig;
  onChange: (field: string, value: string | number | boolean) => void;
}

const desyncModeOptions: { label: string; value: DesyncMode }[] = [
  { label: "Disabled", value: "off" },
  { label: "RST Packets", value: "rst" },
  { label: "FIN Packets", value: "fin" },
  { label: "ACK Packets", value: "ack" },
  { label: "Combo (RST + FIN)", value: "combo" },
  { label: "Full (RST + FIN + ACK)", value: "full" },
];

const desyncModeDescriptions: Record<DesyncMode, string> = {
  off: "No desynchronization",
  rst: "Inject RST packets to disrupt DPI tracking",
  fin: "Inject FIN packets to prematurely close connections",
  ack: "Inject ACK packets to confuse stateful DPI",
  combo: "Use both RST and FIN packets for stronger desync",
  full: "Use RST, FIN, and ACK packets for maximum desync effect",
};

const windowModeOptions: { label: string; value: WindowMode }[] = [
  { label: "Disabled", value: "off" },
  { label: "Zero Window", value: "zero" },
  { label: "Random Window", value: "random" },
  { label: "Oscillate", value: "oscillate" },
  { label: "Escalate", value: "escalate" },
];

const windowModeDescriptions: Record<WindowMode, string> = {
  off: "No window size manipulation",
  zero: "Advertise zero window size to throttle server sending rate",
  random: "Use random window sizes to confuse DPI",
  oscillate: "Alternate between small and large window sizes",
  escalate: "Gradually increase window size over time",
};

export const TcpSettings = ({ config, onChange }: TcpSettingsProps) => {
  return (
    <SettingSection
      title="TCP Configuration"
      description="Configure TCP packet handling"
      icon={<DnsIcon />}
    >
      <Grid container spacing={3}>
        <Grid size={{ xs: 12, md: 6 }}>
          <B4Slider
            label="Connection Bytes Limit"
            value={config.tcp.conn_bytes_limit}
            onChange={(value: number) =>
              onChange("tcp.conn_bytes_limit", value)
            }
            min={1}
            max={100}
            step={1}
            helperText="Bytes to analyze before applying bypass"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <B4Slider
            label="Segment 2 Delay"
            value={config.tcp.seg2delay}
            onChange={(value: number) => onChange("tcp.seg2delay", value)}
            min={0}
            max={1000}
            step={10}
            valueSuffix=" ms"
            helperText="Delay between segments"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <FormControlLabel
            control={
              <Switch
                checked={config.tcp.drop_sack || false}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  onChange("tcp.drop_sack", e.target.checked)
                }
                color="primary"
              />
            }
            label={
              <>
                <Typography variant="body1" fontWeight={500}>
                  Drop SACK Options
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Strip Selective Acknowledgment from TCP headers to confuse
                  stateful DPI
                </Typography>
              </>
            }
          />
        </Grid>
        {/* SYN Fake Settings */}
        <Grid size={{ xs: 12, md: 6 }}>
          <FormControlLabel
            control={
              <Switch
                checked={config.tcp.syn_fake || false}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  onChange("tcp.syn_fake", e.target.checked)
                }
                color="primary"
              />
            }
            label={
              <>
                <Typography variant="body1" fontWeight={500}>
                  SYN Fake Packets
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Send fake SYN packets during TCP handshake (aggressive)
                </Typography>
              </>
            }
          />
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <B4Slider
            label="SYN Fake Payload Length"
            value={config.tcp.syn_fake_len || 0}
            onChange={(value: number) => onChange("tcp.syn_fake_len", value)}
            min={0}
            max={1200}
            step={64}
            disabled={!config.tcp.syn_fake}
            helperText={
              config.tcp.syn_fake
                ? "Fake payload size (0 = use full fake packet)"
                : "Enable SYN Fake to configure length"
            }
          />
        </Grid>
      </Grid>
      <Grid size={{ xs: 12 }}>
        <Divider sx={{ my: 2 }}>
          <Chip label="TCP Window Configuration" size="small" />
        </Divider>
      </Grid>

      <Grid size={{ xs: 12, md: 6 }}>
        <SettingSelect
          label="TCP Window Mode"
          value={config.tcp.win_mode}
          options={windowModeOptions}
          onChange={(e) => onChange("tcp.win_mode", e.target.value as string)}
          helperText={windowModeDescriptions[config.tcp.win_mode]}
        />
      </Grid>

      {config.tcp.win_mode !== "off" && <Grid size={{ xs: 12, md: 6 }}></Grid>}

      <Grid size={{ xs: 12 }}>
        <Divider sx={{ my: 2 }}>
          <Chip label="TCP Desync Configuration" size="small" />
        </Divider>
      </Grid>
      <Grid container spacing={3}>
        <Grid size={{ xs: 12, md: 4 }}>
          <SettingSelect
            label="Desync Mode"
            value={config.tcp.desync_mode}
            options={desyncModeOptions}
            onChange={(e) =>
              onChange("tcp.desync_mode", e.target.value as string)
            }
            helperText={desyncModeDescriptions[config.tcp.desync_mode]}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <B4Slider
            label="Desync TTL"
            value={config.tcp.desync_ttl}
            onChange={(value: number) => onChange("tcp.desync_ttl", value)}
            min={1}
            max={20}
            step={1}
            helperText="TTL value for desync packets"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <B4Slider
            label="Desync Packet Count"
            value={config.tcp.desync_count}
            onChange={(value: number) => onChange("tcp.desync_count", value)}
            min={1}
            max={20}
            step={1}
            helperText="Number of desync packets to send"
          />
        </Grid>
      </Grid>
    </SettingSection>
  );
};
