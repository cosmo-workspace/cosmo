import useUrlState from "@ahooksjs/use-url-state";
import {
  CheckCircleOutlineSharp,
  DescriptionOutlined,
  Edit,
  MoreVert,
  OpenInNewTwoTone,
  RefreshTwoTone,
} from "@mui/icons-material";
import {
  Box,
  Chip,
  FormControl,
  IconButton,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Paper,
  Select,
  SelectChangeEvent,
  Stack,
  Tooltip,
  Typography,
  styled,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import {
  DataGrid,
  GridColDef,
  GridRenderCellParams,
  gridClasses,
} from "@mui/x-data-grid";
import React, { useEffect, useState } from "react";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
import { TemplateDialog } from "../organisms/TemplateDialog";
import { useTemplates as useUserTemplates } from "../organisms/UserModule";
import { useTemplates as useWorkspaceTemplates } from "../organisms/WorkspaceModule";
import { PageTemplate } from "../templates/PageTemplate";

const RotatingRefreshTwoTone = styled(RefreshTwoTone)({
  animation: "rotatingRefresh 1s linear infinite",
  "@keyframes rotatingRefresh": {
    to: {
      transform: "rotate(2turn)",
    },
  },
});

const TemplateMenu: React.VFC<{ template: Template }> = ({ template }) => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  return (
    <>
      <Box>
        <IconButton
          color="inherit"
          onClick={(e) => setAnchorEl(e.currentTarget)}
        >
          <MoreVert fontSize="small" />
        </IconButton>
        <Menu
          anchorEl={anchorEl}
          open={Boolean(anchorEl)}
          onClose={() => setAnchorEl(null)}
        >
          <MenuItem
            onClick={() => {
              // window.open(`/#/event?user=${us.name}`);
            }}
          >
            <ListItemIcon>
              <DescriptionOutlined fontSize="small" />
            </ListItemIcon>
            <ListItemText>
              Show Live Manifest...
              {
                <OpenInNewTwoTone
                  fontSize="inherit"
                  sx={{ position: "relative", top: "0.2em" }}
                />
              }
            </ListItemText>
          </MenuItem>
          <MenuItem
            onClick={() => {
              // window.open(`/#workspace?user=${us.name}`);
            }}
          >
            <ListItemIcon>
              <Edit fontSize="small" />
            </ListItemIcon>
            <ListItemText>
              Edit...
              {
                <OpenInNewTwoTone
                  fontSize="inherit"
                  sx={{ position: "relative", top: "0.2em" }}
                />
              }
            </ListItemText>
          </MenuItem>
        </Menu>
      </Box>
    </>
  );
};

type BaseTemplateDataGridProp = {
  templates: Template[];
  columns: GridColDef[];
};

const BaseTemplateDataGrid: React.FC<BaseTemplateDataGridProp> = ({
  templates,
  columns,
}) => {
  const [open, setOpen] = useState(false);
  const [template, setTemplate] = useState<Template | undefined>(undefined);

  return (
    <>
      <DataGrid
        autoHeight={true}
        rows={templates.map((v) => ({ ...v, id: v.name }))}
        columns={columns}
        getRowHeight={() => "auto"}
        sx={{
          [`& .${gridClasses.cell}`]: {
            py: 1,
          },
        }}
        initialState={{
          pagination: { paginationModel: { pageSize: 10 } },
        }}
        pageSizeOptions={[10, 50, 100]}
        onRowClick={(params) => {
          setTemplate(params.row);
          setOpen(true);
        }}
      />
      <TemplateDialog
        template={template || new Template()}
        open={open}
        onClose={() => setOpen(false)}
      ></TemplateDialog>
    </>
  );
};

type TemplateDataGridProp = {
  templates: Template[];
};

const WorkspaceTemplateDataGrid: React.FC<TemplateDataGridProp> = ({
  templates,
}) => {
  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const columns: GridColDef[] = [
    {
      field: "open",
      headerName: "",
      type: "singleSelect",
      sortable: false,
      disableColumnMenu: true,
      flex: 0.2,
      renderCell: (params: GridRenderCellParams<any, string[]>) => (
        <Tooltip title="Live manifests" placement="top" sx={{ ml: 1 }}>
          <DescriptionOutlined fontSize="small" />
        </Tooltip>
      ),
    },
    {
      field: "name",
      headerName: "Name",
      flex: 2,
    },
    {
      field: "userroles",
      headerName: "Available UserRoles",
      flex: 1,
      minWidth: 100,
      renderCell: (params: GridRenderCellParams<any, string[]>) => (
        <Box
          component="div"
          sx={{
            justifyContent: "flex-end",
            textOverflow: "ellipsis",
            overflow: "hidden",
          }}
        >
          {params.value?.length === 0 ? (
            <Chip color="default" variant="outlined" size="small" label="all" />
          ) : (
            params.value?.map((v, i) => (
              <Chip color="primary" size="small" key={i} label={v} />
            ))
          )}
        </Box>
      ),
    },
    {
      field: "requiredUseraddons",
      headerName: "Dependencies",
      flex: 1,
      minWidth: 100,
      renderCell: (params: GridRenderCellParams<any, string[]>) =>
        params.value && params.value?.length > 0 ? (
          <Stack>
            {params.value?.map((v, i) => (
              <Typography key={i} variant="body2">
                - {v}
              </Typography>
            ))}
          </Stack>
        ) : (
          "-"
        ),
    },
    {
      field: "actions",
      type: "actions",
      getActions: () => [],
      flex: 0.2,
      renderCell: (params: GridRenderCellParams<any, string[]>) => (
        <TemplateMenu template={params.row} />
      ),
    },
  ];

  return <BaseTemplateDataGrid templates={templates} columns={columns} />;
};

const UserAddonDataGrid: React.FC<TemplateDataGridProp> = ({ templates }) => {
  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const columns: GridColDef[] = [
    {
      field: "open",
      headerName: "",
      type: "singleSelect",
      sortable: false,
      disableColumnMenu: true,
      flex: 0.2,
      renderCell: (params: GridRenderCellParams<any, string[]>) =>
        params.row.isDefaultUserAddon ? (
          <Tooltip
            title="Default applied to all users"
            placement="top"
            sx={{ ml: 1 }}
          >
            <CheckCircleOutlineSharp color="success" />
          </Tooltip>
        ) : (
          <Tooltip title="Live manifests" placement="top" sx={{ ml: 1 }}>
            <DescriptionOutlined fontSize="small" />
          </Tooltip>
        ),
    },
    {
      field: "name",
      headerName: "Name",
      flex: 1.3,
      minWidth: 100,
    },
    {
      field: "isClusterScope",
      headerName: "Scope",
      flex: 0.7,
      minWidth: 100,
      renderCell: (params: GridRenderCellParams<any, string[]>) => (
        <Chip
          color="default"
          variant="outlined"
          size="small"
          label={params.row.isClusterScope ? "Cluster" : "Namespace"}
        />
      ),
    },
    {
      field: "userroles",
      headerName: "Available UserRoles",
      flex: 1,
      minWidth: 100,
      renderCell: (params: GridRenderCellParams<any, string[]>) => (
        <Box
          component="div"
          sx={{
            justifyContent: "flex-end",
            textOverflow: "ellipsis",
            overflow: "hidden",
          }}
        >
          {params.value?.length === 0 ? (
            <Chip color="default" variant="outlined" size="small" label="all" />
          ) : (
            params.value?.map((v, i) => (
              <Chip color="primary" size="small" key={i} label={v} />
            ))
          )}
        </Box>
      ),
    },
    {
      field: "requiredUseraddons",
      headerName: "Dependencies",
      flex: 1,
      minWidth: 100,
      renderCell: (params: GridRenderCellParams<any, string[]>) =>
        params.value && params.value?.length > 0 ? (
          <Stack>
            {params.value?.map((v, i) => (
              <Typography key={i} variant="body2">
                - {v}
              </Typography>
            ))}
          </Stack>
        ) : (
          "-"
        ),
    },
    {
      field: "actions",
      type: "actions",
      getActions: () => [],
      flex: 0.2,
      renderCell: (params: GridRenderCellParams<any, string[]>) => (
        <TemplateMenu template={params.row} />
      ),
    },
  ];
  return <BaseTemplateDataGrid templates={templates} columns={columns} />;
};

const TemplateList: React.VFC = () => {
  console.log("TemplateList");
  const [isLoading, setIsLoading] = React.useState(false);
  const [urlParam, setUrlParam] = useUrlState(
    { type: "workspace" },
    {
      stringifyOptions: { skipEmptyString: true },
    }
  );
  const templateType = urlParam.type;
  const setTemplateType = (type: string) => {
    setUrlParam({ type });
  };
  if (!["workspace", "useraddon"].includes(templateType))
    throw new Error("Invalid template type");

  const { templates: workspaceTemplates, getTemplates } =
    useWorkspaceTemplates();
  const { templates: useraddons, getUserAddonTemplates } = useUserTemplates();

  const refreshTemplates = () => {
    switch (templateType) {
      case "workspace":
        getTemplates({ withRaw: true });
        break;
      case "useraddon":
        getUserAddonTemplates({ withRaw: true });
        break;
    }
  };

  useEffect(refreshTemplates, [templateType]);

  return (
    <>
      <Paper sx={{ minWidth: 320, maxWidth: 1200, px: 2, py: 1 }}>
        <Stack direction="row" alignItems="center" spacing={2}>
          <Box sx={{ flexGrow: 1 }} />
          <FormControl sx={{ m: 1, minWidth: 200 }} size="small">
            <Tooltip title="Choose Template Type" placement="top">
              <Select
                labelId="demo-select-small-label"
                id="demo-select-small"
                value={templateType}
                onChange={(event: SelectChangeEvent) => {
                  setTemplateType(event.target.value as string);
                }}
              >
                <MenuItem value="workspace">WorkspaceTemplate</MenuItem>
                <MenuItem value="useraddon">UserAddon</MenuItem>
              </Select>
            </Tooltip>
          </FormControl>
          <Tooltip title="Refresh" placement="top">
            <IconButton
              color="inherit"
              onClick={() => {
                setIsLoading(true);
                setTimeout(() => {
                  setIsLoading(false);
                }, 1000);
                if (!isLoading) refreshTemplates();
              }}
            >
              {isLoading ? <RotatingRefreshTwoTone /> : <RefreshTwoTone />}
            </IconButton>
          </Tooltip>
        </Stack>
      </Paper>
      {templateType === "workspace" && (
        <WorkspaceTemplateDataGrid templates={workspaceTemplates} />
      )}
      {templateType === "useraddon" && (
        <UserAddonDataGrid templates={useraddons} />
      )}
    </>
  );
};

export const TemplatePage: React.VFC = () => {
  console.log("TemplatePage");

  return (
    <PageTemplate title="Templates">
      <TemplateList />
    </PageTemplate>
  );
};
