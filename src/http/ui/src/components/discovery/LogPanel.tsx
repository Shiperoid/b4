import { useEffect, useRef, useState } from "react";
import {
  Box,
  IconButton,
  Typography,
  Stack,
  Tooltip,
  Paper,
} from "@mui/material";
import { ExpandIcon, CollapseIcon, ClearIcon, LogsIcon } from "@b4.icons";
import { colors } from "@design";
import { useDiscoveryLogs } from "@b4.discovery";
import { B4Badge } from "@b4.elements";

interface DiscoveryLogPanelProps {
  running: boolean;
}

export const DiscoveryLogPanel = ({ running }: DiscoveryLogPanelProps) => {
  const { logs, connected, clearLogs } = useDiscoveryLogs();
  const [expanded, setExpanded] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const hasAutoExpanded = useRef(false);

  useEffect(() => {
    if (running && logs.length > 0 && !hasAutoExpanded.current) {
      setExpanded(true);
      hasAutoExpanded.current = true;
    }
    if (!running) {
      hasAutoExpanded.current = false;
    }
  }, [running, logs.length]);

  useEffect(() => {
    if (scrollRef.current && expanded) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs, expanded]);

  if (!running && logs.length === 0) return null;

  return (
    <Paper
      elevation={0}
      sx={{
        bgcolor: colors.background.paper,
        border: `1px solid ${colors.border.default}`,
        borderRadius: 2,
        overflow: "hidden",
      }}
    >
      {/* Header */}
      <Stack
        direction="row"
        alignItems="center"
        justifyContent="space-between"
        sx={{
          p: 2,
          bgcolor: colors.accent.primary,
          cursor: "pointer",
        }}
        onClick={() => setExpanded((e) => !e)}
      >
        <Stack direction="row" alignItems="center" spacing={1.5}>
          <LogsIcon sx={{ fontSize: 20, color: colors.secondary }} />
          <Typography variant="h6" sx={{ color: colors.text.primary }}>
            Discovery Logs
          </Typography>
          <Box
            sx={{
              width: 16,
              height: 16,
              borderRadius: "50%",
              bgcolor: connected ? colors.secondary : colors.text.disabled,
            }}
          />
          {logs.length > 0 && (
            <B4Badge variant="filled" label={`${logs.length} lines`} />
          )}
        </Stack>
        <Stack direction="row" alignItems="center" spacing={1}>
          {logs.length > 0 && (
            <Tooltip title="Clear logs">
              <IconButton
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  clearLogs();
                }}
                sx={{ color: colors.text.secondary }}
              >
                <ClearIcon fontSize="small" />
              </IconButton>
            </Tooltip>
          )}
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();
              setExpanded((prev) => !prev);
            }}
            sx={{ color: colors.text.secondary }}
          >
            {expanded ? <CollapseIcon /> : <ExpandIcon />}
          </IconButton>
        </Stack>
      </Stack>

      {/* Log content - simple div, not Box */}
      {expanded && (
        <div
          ref={scrollRef}
          style={{
            height: 150,
            overflowY: "auto",
            backgroundColor: colors.background.dark,
            fontFamily: "monospace",
            fontSize: 12,
            padding: 16,
          }}
        >
          {logs.length === 0 ? (
            <Typography
              sx={{ color: colors.text.disabled, fontStyle: "italic" }}
            >
              Waiting for discovery logs...
            </Typography>
          ) : (
            logs.map((line, i) => (
              <div
                key={i}
                style={{
                  color: getLogColor(line),
                  whiteSpace: "pre-wrap",
                  wordBreak: "break-word",
                  lineHeight: 1.6,
                }}
              >
                {line}
              </div>
            ))
          )}
        </div>
      )}
    </Paper>
  );
};

function getLogColor(line: string): string {
  const lower = line.toLowerCase();
  if (lower.includes("success") || line.includes("✓") || lower.includes("best"))
    return colors.secondary;
  if (lower.includes("failed") || line.includes("✗") || lower.includes("fail"))
    return colors.primary;
  if (lower.includes("phase")) return colors.text.secondary;
  return colors.text.primary;
}
