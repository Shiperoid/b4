import React from "react";
import {
  AppBar,
  Box,
  Container,
  CssBaseline,
  IconButton,
  Paper,
  Stack,
  Toolbar,
  Typography,
  ThemeProvider,
  createTheme,
  Switch,
  FormControlLabel,
  TextField,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";

const theme = createTheme({
  palette: {
    mode: "dark",
    primary: { main: "#9E1C60" },
    secondary: { main: "#F5AD18" },
    info: { main: "#811844" },
    error: { main: "#561530" },
    background: { default: "#1a0e15", paper: "#1f1218" },
    text: { primary: "#ffe8f4", secondary: "#f8d7e9" },
  },
  components: {
    MuiAppBar: {
      styleOverrides: {
        root: {
          background:
            "linear-gradient(90deg, #561530 0%, #811844 35%, #9E1C60 70%, #F5AD18 100%)",
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: "none",
          borderColor: "rgba(245, 173, 24, 0.24)",
        },
      },
    },
  },
  typography: {
    fontFamily:
      'system-ui, -apple-system, "Segoe UI", Roboto, Ubuntu, "Helvetica Neue", Arial',
  },
});

export default function App() {
  const [lines, setLines] = React.useState<string[]>([]);
  const [paused, setPaused] = React.useState(false);
  const [filter, setFilter] = React.useState("");
  const logRef = React.useRef<HTMLDivElement | null>(null);

  React.useEffect(() => {
    const ws = new WebSocket(
      (location.protocol === "https:" ? "wss://" : "ws://") +
        location.host +
        "/api/ws/logs"
    );
    ws.onmessage = (ev) => {
      if (!paused) setLines((prev) => [...prev.slice(-999), String(ev.data)]);
    };
    ws.onerror = () => setLines((prev) => [...prev, "[WS ERROR]"]);
    return () => ws.close();
  }, [paused]);

  React.useEffect(() => {
    const el = logRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [lines]);

  const filtered = React.useMemo(() => {
    const f = filter.trim().toLowerCase();
    return f ? lines.filter((l) => l.toLowerCase().includes(f)) : lines;
  }, [lines, filter]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppBar position="sticky" elevation={1}>
        <Toolbar>
          <Typography variant="h6" sx={{ flexGrow: 1 }}>
            B4
          </Typography>
          <FormControlLabel
            sx={{ mr: 2 }}
            control={
              <Switch
                checked={paused}
                onChange={(e) => setPaused(e.target.checked)}
              />
            }
            label="Pause"
          />
          <IconButton color="inherit" onClick={() => setLines([])}>
            <RefreshIcon />
          </IconButton>
        </Toolbar>
      </AppBar>
      <Container maxWidth="lg" sx={{ py: 3 }}>
        <Stack spacing={2}>
          <TextField
            size="small"
            label="Filter"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            fullWidth
          />
          <Paper variant="outlined" sx={{ p: 1.5 }}>
            <Box
              ref={logRef}
              sx={{
                height: 480,
                overflowY: "auto",
                fontFamily:
                  'ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace',
                fontSize: 13,
                lineHeight: 1.4,
                whiteSpace: "pre-wrap",
                wordBreak: "break-word",
                backgroundColor: "#0f0a0e",
                color: "text.primary",
                borderRadius: 1,
              }}
            >
              {filtered.map((l, i) => (
                <Typography
                  key={i}
                  component="div"
                  sx={{ fontFamily: "inherit", fontSize: "inherit" }}
                >
                  {l}
                </Typography>
              ))}
            </Box>
          </Paper>
        </Stack>
      </Container>
    </ThemeProvider>
  );
}
