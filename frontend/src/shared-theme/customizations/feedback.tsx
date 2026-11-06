import { Theme, alpha, Components } from '@mui/material/styles';
import { gray } from '../themePrimitives';

/* eslint-disable import/prefer-default-export */
export const feedbackCustomizations: Components<Theme> = {
  MuiAlert: {
    styleOverrides: {
      // the soft look of the template, in the color of the alert's severity
      // (it used to be orange for every severity, so errors read as warnings)
      root: ({ theme, ownerState }) => {
        const main = theme.palette[ownerState.severity ?? 'success'].main;
        return {
          borderRadius: 10,
          backgroundColor: alpha(main, 0.08),
          color: (theme.vars || theme).palette.text.primary,
          border: `1px solid ${alpha(main, 0.35)}`,
          '& .MuiAlert-icon': {
            color: main,
          },
          ...theme.applyStyles('dark', {
            backgroundColor: alpha(main, 0.12),
            border: `1px solid ${alpha(main, 0.4)}`,
          }),
        };
      },
    },
  },
  MuiDialog: {
    styleOverrides: {
      root: ({ theme }) => ({
        '& .MuiDialog-paper': {
          borderRadius: '10px',
          border: '1px solid',
          borderColor: (theme.vars || theme).palette.divider,
        },
      }),
    },
  },
  MuiLinearProgress: {
    styleOverrides: {
      root: ({ theme }) => ({
        height: 8,
        borderRadius: 8,
        backgroundColor: gray[200],
        ...theme.applyStyles('dark', {
          backgroundColor: gray[800],
        }),
      }),
    },
  },
};