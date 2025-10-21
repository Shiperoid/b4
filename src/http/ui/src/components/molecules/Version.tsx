import { Box, Link, Typography } from "@mui/material";
import { colors } from "../../Theme";
import GitHubIcon from "@mui/icons-material/GitHub";

export default function Version() {
  return (
    <Box
      sx={{
        py: 2,
        textAlign: "center",
      }}
    >
      <Typography
        variant="caption"
        sx={{
          color: colors.secondary,
        }}
      >
        <Link
          href="https://github.com/daniellavrushin/b4"
          target="_blank"
          rel="noopener noreferrer"
        >
          <GitHubIcon
            sx={{
              verticalAlign: "text-bottom",
              mr: 0.5,
              fontSize: "1rem",
            }}
          />
        </Link>
        Version 1.0.0
      </Typography>
    </Box>
  );
}
