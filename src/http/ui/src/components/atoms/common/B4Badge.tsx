import { Chip, ChipProps } from "@mui/material";
import { colors } from "@design";

type BadgeVariant = "primary" | "secondary";

interface B4BadgeProps extends Omit<ChipProps, "color" | "variant"> {
  badgeVariant?: BadgeVariant;
}

const variantStyles: Record<BadgeVariant, object> = {
  primary: {
    bgcolor: colors.accent.primary,
    borderColor: colors.primary,
  },
  secondary: {
    bgcolor: `${colors.tertiary}`,
    borderColor: colors.tertiary,
  },
};

export const B4Badge: React.FC<B4BadgeProps> = ({
  badgeVariant = "primary",
  sx,
  ...props
}) => (
  <Chip
    size="small"
    sx={{
      ...variantStyles[badgeVariant],
      ...sx,
    }}
    {...props}
  />
);
