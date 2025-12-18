import { useState, useEffect } from "react";
import {
  Grid,
  Box,
  Typography,
  Chip,
  CircularProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Checkbox,
  Paper,
} from "@mui/material";
import { DeviceUnknowIcon, RefreshIcon } from "@b4.icons";
import { B4Config } from "@models/config";
import { colors } from "@design";
import {
  B4Section,
  B4Switch,
  B4Alert,
  B4TooltipButton,
  B4Badge,
} from "@b4.elements";

interface DeviceInfo {
  mac: string;
  ip: string;
  hostname: string;
  vendor: string;
  country: string;
}

interface DevicesResponse {
  available: boolean;
  source?: string;
  devices: DeviceInfo[];
}

interface DevicesSettingsProps {
  config: B4Config;
  onChange: (field: string, value: boolean | string | string[]) => void;
}

export const DevicesSettings = ({ config, onChange }: DevicesSettingsProps) => {
  const [devices, setDevices] = useState<DeviceInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [available, setAvailable] = useState(false);
  const [source, setSource] = useState<string>("");

  const selectedMacs = config.queue.devices?.mac || [];
  const enabled = config.queue.devices?.enabled || false;
  const wisb = config.queue.devices?.wisb || false;

  const loadDevices = async () => {
    setLoading(true);
    try {
      const resp = await fetch("/api/devices");
      const data = (await resp.json()) as DevicesResponse;
      setAvailable(data.available);
      setSource(data.source || "");
      setDevices(data.devices || []);
    } catch (err) {
      console.error("Failed to load devices:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadDevices();
  }, []);

  const handleMacToggle = (mac: string) => {
    const current = [...selectedMacs];
    const index = current.indexOf(mac);
    if (index === -1) {
      current.push(mac);
    } else {
      current.splice(index, 1);
    }
    onChange("queue.devices.mac", current);
  };

  const isSelected = (mac: string) => selectedMacs.includes(mac);

  return (
    <B4Section
      title="Device Filtering"
      description="Filter traffic by source device MAC address"
      icon={<DeviceUnknowIcon />}
    >
      <Grid container spacing={2}>
        <Grid size={{ xs: 12 }}>
          <Box sx={{ display: "flex", gap: 3, alignItems: "center" }}>
            <B4Switch
              label="Enable Device Filtering"
              checked={enabled}
              onChange={(checked) => onChange("queue.devices.enabled", checked)}
              description="Only process traffic from selected devices"
            />
            <B4Switch
              label="Invert Selection (Blacklist)"
              checked={wisb}
              onChange={(checked) => onChange("queue.devices.wisb", checked)}
              description={
                wisb ? "Block selected devices" : "Allow only selected devices"
              }
              disabled={!enabled}
            />
          </Box>
        </Grid>

        {enabled && (
          <>
            <Grid size={{ xs: 12 }}>
              <B4Alert severity={wisb ? "warning" : "info"}>
                {wisb
                  ? "Blacklist mode: Selected devices will be EXCLUDED from DPI bypass"
                  : "Whitelist mode: Only selected devices will use DPI bypass"}
              </B4Alert>
            </Grid>

            {!available ? (
              <Grid size={{ xs: 12 }}>
                <B4Alert severity="warning">
                  DHCP lease source not detected. Device discovery unavailable.
                  You can still manually add MAC addresses below.
                </B4Alert>
              </Grid>
            ) : (
              <Grid size={{ xs: 12 }}>
                <Box
                  sx={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    mb: 1,
                  }}
                >
                  <Typography variant="subtitle2">
                    Available Devices
                    {source && (
                      <Chip
                        label={source}
                        size="small"
                        sx={{
                          ml: 1,
                          bgcolor: colors.accent.secondary,
                          color: colors.secondary,
                        }}
                      />
                    )}
                  </Typography>
                  <B4TooltipButton
                    title="Refresh devices"
                    icon={
                      loading ? <CircularProgress size={18} /> : <RefreshIcon />
                    }
                    onClick={() => void loadDevices()}
                  />
                </Box>

                <TableContainer
                  component={Paper}
                  sx={{
                    bgcolor: colors.background.paper,
                    border: `1px solid ${colors.border.default}`,
                    maxHeight: 300,
                  }}
                >
                  <Table size="small" stickyHeader>
                    <TableHead>
                      <TableRow>
                        <TableCell
                          padding="checkbox"
                          sx={{ bgcolor: colors.background.dark }}
                        >
                          <Checkbox
                            color="secondary"
                            indeterminate={
                              selectedMacs.length > 0 &&
                              selectedMacs.length < devices.length
                            }
                            checked={
                              devices.length > 0 &&
                              selectedMacs.length === devices.length
                            }
                            onChange={(e) => {
                              if (e.target.checked) {
                                onChange(
                                  "queue.devices.mac",
                                  devices.map((d) => d.mac)
                                );
                              } else {
                                onChange("queue.devices.mac", []);
                              }
                            }}
                          />
                        </TableCell>
                        <TableCell
                          sx={{
                            bgcolor: colors.background.dark,
                            color: colors.text.secondary,
                          }}
                        >
                          MAC Address
                        </TableCell>
                        <TableCell
                          sx={{
                            bgcolor: colors.background.dark,
                            color: colors.text.secondary,
                          }}
                        >
                          IP
                        </TableCell>
                        <TableCell
                          sx={{
                            bgcolor: colors.background.dark,
                            color: colors.text.secondary,
                          }}
                        >
                          Vendor
                        </TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {devices.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={4} align="center">
                            {loading
                              ? "Loading devices..."
                              : "No devices found"}
                          </TableCell>
                        </TableRow>
                      ) : (
                        devices.map((device) => (
                          <TableRow
                            key={device.mac}
                            hover
                            onClick={() => handleMacToggle(device.mac)}
                            sx={{ cursor: "pointer" }}
                          >
                            <TableCell padding="checkbox">
                              <Checkbox
                                checked={isSelected(device.mac)}
                                color="secondary"
                              />
                            </TableCell>
                            <TableCell
                              sx={{
                                fontFamily: "monospace",
                                fontSize: "0.85rem",
                              }}
                            >
                              {device.mac}
                            </TableCell>
                            <TableCell
                              sx={{
                                fontFamily: "monospace",
                                fontSize: "0.85rem",
                              }}
                            >
                              {device.ip}
                            </TableCell>
                            <TableCell>
                              {device.vendor ? (
                                <B4Badge
                                  label={device.vendor}
                                  color="primary"
                                  variant={
                                    isSelected(device.mac)
                                      ? "filled"
                                      : "outlined"
                                  }
                                />
                              ) : (
                                <Typography
                                  variant="caption"
                                  color="text.secondary"
                                >
                                  Unknown
                                </Typography>
                              )}
                            </TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </TableContainer>
              </Grid>
            )}
          </>
        )}
      </Grid>
    </B4Section>
  );
};
