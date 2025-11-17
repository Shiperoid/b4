import React, { useState, useEffect } from "react";
import {
  Box,
  Button,
  Stack,
  Typography,
  LinearProgress,
  Alert,
  Paper,
  Divider,
  Grid,
  Chip,
  IconButton,
} from "@mui/material";
import {
  PlayArrow as StartIcon,
  Stop as StopIcon,
  Refresh as RefreshIcon,
  Add as AddIcon,
} from "@mui/icons-material";
import { button_secondary, colors } from "@design";
import { TestResultCard } from "@molecules/check/ResultCard";
import { TestStatus } from "@atoms/check/Badge";
import { useConfigLoad } from "@hooks/useConfig";
import SettingTextField from "@atoms/common/B4TextField";
import { useTestDomains } from "@hooks/useTestDomains";

interface TestResult {
  domain: string;
  status: TestStatus;
  duration: number;
  speed: number;
  bytes_read: number;
  error?: string;
  timestamp: string;
  is_baseline: boolean;
  improvement: number;
  status_code: number;
}

interface TestSuite {
  id: string;
  status: TestStatus;
  start_time: string;
  end_time: string;
  total_checks: number;
  completed_checks: number;
  successful_checks: number;
  failed_checks: number;
  results: TestResult[];
  summary: {
    average_speed: number;
    average_improvement: number;
    fastest_domain: string;
    slowest_domain: string;
    success_rate: number;
  };
}

interface TestRunnerProps {
  onStart?: () => void;
  onComplete?: (suite: TestSuite) => void;
}

export const TestRunner: React.FC<TestRunnerProps> = ({
  onStart,
  onComplete,
}) => {
  const [running, setRunning] = useState(false);
  const [testId, setTestId] = useState<string | null>(null);
  const [suite, setSuite] = useState<TestSuite | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { config } = useConfigLoad();
  const { domains, addDomain, removeDomain, clearDomains, resetToDefaults } =
    useTestDomains();
  const [newDomain, setNewDomain] = useState("");

  // Poll for test status
  useEffect(() => {
    if (!testId || !running) return;

    const fetchStatus = async () => {
      try {
        const response = await fetch(`/api/check/status?id=${testId}`);
        if (!response.ok) {
          throw new Error("Failed to fetch test status");
        }

        const data: TestSuite = (await response.json()) as TestSuite;
        setSuite(data);

        if (
          data.status === "complete" ||
          data.status === "failed" ||
          data.status === "canceled"
        ) {
          setRunning(false);
          if (onComplete) {
            onComplete(data);
          }
        }
      } catch (err) {
        console.error("Failed to fetch test status:", err);
        setError(err instanceof Error ? err.message : "Unknown error");
        setRunning(false);
      }
    };

    const interval = setInterval(() => {
      void fetchStatus();
    }, 1000);

    return () => clearInterval(interval);
  }, [testId, running, onComplete]);

  const startTest = async () => {
    if (domains.length === 0) {
      setError("Add at least one domain to test");
      return;
    }

    setError(null);
    setRunning(true);
    setSuite(null);

    if (onStart) {
      onStart();
    }

    try {
      const timeout = (config?.system.checker.timeout || 15) * 1e9;
      const maxConcurrent = config?.system.checker.max_concurrent || 5;

      const response = await fetch("/api/check/start", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          timeout: timeout,
          max_concurrent: maxConcurrent,
          domains: domains,
        }),
      });

      if (!response.ok) {
        const text = await response.text();
        throw new Error(text || "Failed to start test");
      }

      const data = (await response.json()) as { id: string };
      setTestId(data.id);
    } catch (err) {
      console.error("Failed to start test:", err);
      setError(err instanceof Error ? err.message : "Failed to start test");
      setRunning(false);
    }
  };

  const cancelTest = async () => {
    if (!testId) return;

    try {
      await fetch(`/api/check/cancel?id=${testId}`, { method: "DELETE" });
      setRunning(false);
    } catch (err) {
      console.error("Failed to cancel test:", err);
    }
  };

  const resetTest = () => {
    setTestId(null);
    setSuite(null);
    setError(null);
    setRunning(false);
  };

  const progress = suite
    ? (suite.completed_checks / suite.total_checks) * 100
    : 0;

  return (
    <Stack spacing={3}>
      {/* Control Panel */}
      <Paper
        elevation={0}
        sx={{
          p: 3,
          bgcolor: colors.background.paper,
          border: `1px solid ${colors.border.default}`,
          borderRadius: 2,
        }}
      >
        <Stack spacing={2}>
          {/* Header with actions */}
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
            }}
          >
            <Typography variant="h6" sx={{ color: colors.text.primary }}>
              DPI Bypass Test Suite
            </Typography>
            <Stack direction="row" spacing={1}>
              {!running && !suite && (
                <Button
                  variant="contained"
                  startIcon={<StartIcon />}
                  onClick={() => {
                    void startTest();
                  }}
                  disabled={domains.length === 0}
                  sx={{
                    bgcolor: colors.secondary,
                    "&:hover": { bgcolor: colors.primary },
                    "&:disabled": {
                      bgcolor: colors.accent.secondary,
                      color: colors.text.secondary,
                    },
                  }}
                >
                  Start Test
                </Button>
              )}
              {running && (
                <Button
                  variant="outlined"
                  startIcon={<StopIcon />}
                  onClick={() => {
                    void cancelTest();
                  }}
                  sx={{
                    borderColor: colors.quaternary,
                    color: colors.quaternary,
                  }}
                >
                  Cancel
                </Button>
              )}
              {suite && !running && (
                <Button
                  variant="outlined"
                  startIcon={<RefreshIcon />}
                  onClick={resetTest}
                  sx={{
                    borderColor: colors.secondary,
                    color: colors.secondary,
                  }}
                >
                  New Test
                </Button>
              )}
            </Stack>
          </Box>

          {error && <Alert severity="error">{error}</Alert>}

          {/* Domain Management Section */}
          <Box>
            <Stack
              direction="row"
              alignItems="center"
              justifyContent="space-between"
              sx={{ mb: 1 }}
            >
              <Typography
                variant="subtitle2"
                sx={{ color: colors.text.primary }}
              >
                Domains to Test
              </Typography>
              <Stack direction="row" spacing={1}>
                <Button
                  size="small"
                  onClick={resetToDefaults}
                  disabled={running}
                  sx={{ ...button_secondary, textTransform: "none" }}
                >
                  Reset to Defaults
                </Button>
                <Button
                  size="small"
                  onClick={clearDomains}
                  disabled={running || domains.length === 0}
                  sx={{ ...button_secondary, textTransform: "none" }}
                >
                  Clear All
                </Button>
              </Stack>
            </Stack>

            <Grid container spacing={2}>
              <Grid size={{ sm: 12, md: 6 }}>
                {/* Domain Input */}
                <Box
                  sx={{
                    display: "flex",
                    gap: 1,
                    pb: 2,
                    width: "100%",
                    alignItems: "flex-start",
                  }}
                >
                  <SettingTextField
                    fullWidth
                    label="Add domain"
                    value={newDomain}
                    onChange={(e) => setNewDomain(e.target.value)}
                    onKeyDown={(e) => {
                      if (
                        e.key === "Enter" ||
                        e.key === "," ||
                        e.key === "Tab"
                      ) {
                        e.preventDefault();
                        addDomain(newDomain);
                        setNewDomain("");
                      }
                    }}
                    placeholder="youtube.com"
                    disabled={running}
                    helperText="Press Enter or comma to add"
                  />
                  <IconButton
                    onClick={() => {
                      addDomain(newDomain);
                      setNewDomain("");
                    }}
                    disabled={running || !newDomain.trim()}
                    sx={{
                      bgcolor: colors.accent.secondary,
                      color: colors.secondary,
                      "&:hover": {
                        bgcolor: colors.accent.secondaryHover,
                      },
                    }}
                  >
                    <AddIcon />
                  </IconButton>
                </Box>
              </Grid>
              <Grid size={{ sm: 12, md: 6 }}>
                {/* Domain Chips */}
                <Box
                  sx={{
                    display: "flex",
                    flexWrap: "wrap",
                    gap: 1,
                    p: 2,
                    width: "100%",
                    border: `1px solid ${colors.border.default}`,
                    borderRadius: 1,
                    bgcolor: colors.background.dark,
                  }}
                >
                  {domains.length === 0 ? (
                    <Typography
                      variant="body2"
                      sx={{
                        color: colors.text.secondary,
                        width: "100%",
                        textAlign: "center",
                      }}
                    >
                      No domains added. Add domains above or click "Reset to
                      Defaults"
                    </Typography>
                  ) : (
                    domains.map((domain) => (
                      <Chip
                        size="small"
                        key={domain}
                        label={domain}
                        onDelete={() => removeDomain(domain)}
                        disabled={running}
                        sx={{
                          bgcolor: colors.accent.primary,
                          color: colors.secondary,
                          "& .MuiChip-deleteIcon": {
                            color: colors.secondary,
                          },
                        }}
                      />
                    ))
                  )}
                </Box>
              </Grid>
            </Grid>
          </Box>

          {/* Progress indicator */}
          {running && suite && (
            <Box>
              <Box
                sx={{
                  display: "flex",
                  justifyContent: "space-between",
                  mb: 1,
                }}
              >
                <Typography variant="body2" color="text.secondary">
                  Testing {suite.completed_checks} of {suite.total_checks}{" "}
                  domains
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {progress.toFixed(0)}%
                </Typography>
              </Box>
              <LinearProgress
                variant="determinate"
                value={progress}
                sx={{
                  height: 8,
                  borderRadius: 4,
                  bgcolor: colors.background.dark,
                  "& .MuiLinearProgress-bar": {
                    bgcolor: colors.secondary,
                    borderRadius: 4,
                  },
                }}
              />
            </Box>
          )}
        </Stack>
      </Paper>

      {/* Summary */}
      {suite && !running && suite.status === "complete" && (
        <Paper
          elevation={0}
          sx={{
            p: 3,
            bgcolor: colors.background.paper,
            border: `1px solid ${colors.border.default}`,
            borderRadius: 2,
          }}
        >
          <Typography variant="h6" sx={{ mb: 2, color: colors.text.primary }}>
            Test Summary
          </Typography>
          <Divider sx={{ mb: 2, borderColor: colors.border.default }} />
          <Grid container spacing={2}>
            <Grid size={{ xs: 12, sm: 6, md: 3 }}>
              <Box sx={{ textAlign: "center" }}>
                <Typography variant="h4" color="primary">
                  {suite.successful_checks}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Successful
                </Typography>
              </Box>
            </Grid>
            <Grid size={{ xs: 12, sm: 6, md: 3 }}>
              <Box sx={{ textAlign: "center" }}>
                <Typography variant="h4" color="error">
                  {suite.failed_checks}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Failed
                </Typography>
              </Box>
            </Grid>
            <Grid size={{ xs: 12, sm: 6, md: 3 }}>
              <Box sx={{ textAlign: "center" }}>
                <Typography variant="h4" color="secondary">
                  {suite.summary.success_rate.toFixed(1)}%
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Success Rate
                </Typography>
              </Box>
            </Grid>
            <Grid size={{ xs: 12, sm: 6, md: 3 }}>
              <Box sx={{ textAlign: "center" }}>
                <Typography variant="h4" sx={{ color: colors.secondary }}>
                  {(suite.summary.average_speed / 1024 / 1024).toFixed(2)}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Avg Speed (MB/s)
                </Typography>
              </Box>
            </Grid>
          </Grid>
        </Paper>
      )}

      {/* Results Grid */}
      {suite?.results && suite.results.length > 0 && (
        <Box>
          <Typography variant="h6" sx={{ mb: 2, color: colors.text.primary }}>
            Test Results
          </Typography>
          <Grid container spacing={2}>
            {suite.results.map((result) => (
              <Grid key={result.domain} size={{ xs: 12, md: 6, lg: 4 }}>
                <TestResultCard
                  domain={result.domain}
                  status={result.status}
                  duration={result.duration / 1000000}
                  speed={result.speed}
                  improvement={result.improvement}
                  error={result.error}
                  status_code={result.status_code}
                />
              </Grid>
            ))}
          </Grid>
        </Box>
      )}
    </Stack>
  );
};
