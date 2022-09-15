import { createTheme } from "@mui/material/styles";

let theme = createTheme({
    palette: {
        primary: {
            light: '#63ccff',
            main: '#009be5',
            dark: '#006db3',
        },
        secondary: {
            light: '#CC3333',
            main: '#bb1c2a',
            dark: '#951621',
        },
    },
    typography: {
        h1: {
            fontWeight: 500,
            fontSize: 30,
            letterSpacing: 0.5,
        },
        h2: {
            fontWeight: 500,
            fontSize: 29,
            letterSpacing: 0.5,
        },
        h3: {
            fontWeight: 500,
            fontSize: 27,
            letterSpacing: 0.5,
        },
        h4: {
            fontWeight: 500,
            fontSize: 25,
            letterSpacing: 0.5,
        },
        h5: {
            fontWeight: 500,
            fontSize: 20,
            letterSpacing: 0.5,
        },
        h6: {
            fontWeight: 500,
            fontSize: 18,
            letterSpacing: 0.5,
        },
    },
    shape: {
        borderRadius: 8,
    },
    components: {
        MuiTab: {
            defaultProps: {
                disableRipple: true,
            },
        },
        MuiPaper: {
            variants: [
                {
                    props: { variant: 'paper_content' },
                    style: {
                        maxWidth: 1200,
                        margin: 'auto',
                        overflow: 'hidden',
                    },
                },
            ],
        },
    },
    mixins: {
        toolbar: {
            minHeight: 48,
        },
    },
})

theme = {
    ...theme,
    components: {
        MuiDrawer: {
            styleOverrides: {
                paper: {
                    backgroundColor: '#081627',
                },
            },
        },
        MuiButton: {
            styleOverrides: {
                root: {
                    textTransform: 'none',
                },
                contained: {
                    boxShadow: 'none',
                    '&:active': {
                        boxShadow: 'none',
                    },
                },
            },
        },
        MuiTabs: {
            styleOverrides: {
                root: {
                    marginLeft: theme.spacing(1),
                },
                indicator: {
                    height: 3,
                    borderTopLeftRadius: 3,
                    borderTopRightRadius: 3,
                    backgroundColor: theme.palette.common.white,
                },
            },
        },
        MuiTab: {
            styleOverrides: {
                root: {
                    textTransform: 'none',
                    margin: '0 16px',
                    minWidth: 0,
                    padding: 0,
                    [ theme.breakpoints.up('md') ]: {
                        padding: 0,
                        minWidth: 0,
                    },
                },
            },
        },
        MuiIconButton: {
            styleOverrides: {
                root: {
                    padding: theme.spacing(1),
                },
            },
        },
        MuiTooltip: {
            styleOverrides: {
                tooltip: {
                    borderRadius: 4,
                },
            },
        },
        MuiDivider: {
            styleOverrides: {
                root: {
                    backgroundColor: 'rgb(255,255,255,0.15)',
                },
            },
        },
        MuiListItemButton: {
            styleOverrides: {
                root: {
                    '&.Mui-selected': {
                        color: '#4fc3f7',
                    },
                },
            },
        },
        MuiListItemText: {
            styleOverrides: {
                primary: {
                    fontSize: 14,
                    fontWeight: theme.typography.fontWeightMedium,
                },
            },
        },
        MuiListItemIcon: {
            styleOverrides: {
                root: {
                    color: 'inherit',
                    minWidth: 'auto',
                    marginRight: theme.spacing(2),
                    '& svg': {
                        fontSize: 20,
                    },
                },
            },
        },
        MuiAvatar: {
            styleOverrides: {
                root: {
                    width: 32,
                    height: 32,
                },
            },
        },
    },
};

export default theme;