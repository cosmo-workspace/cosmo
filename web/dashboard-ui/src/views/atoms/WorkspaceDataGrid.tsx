import { Timestamp } from "@bufbuild/protobuf";
import { SxProps, useMediaQuery, useTheme } from "@mui/material";
import {
  DataGrid,
  DataGridProps,
  GridColDef,
  GridRenderCellParams,
  gridClasses,
} from "@mui/x-data-grid";
import React from "react";
import { Workspace } from "../../proto/gen/dashboard/v1alpha1/workspace_pb";
import { WorkspaceInfoDialogContext } from "../organisms/WorkspaceInfoDialog";
import { StatusChip } from "./StatusChip";

export type WorkspaceDataGridProp = {
  workspaces: Workspace[];
  maxHeight?: number;
  sx?: SxProps;
  dataGridProps?: Omit<DataGridProps, "columns" | "sx">;
};

export const WorkspaceDataGrid: React.FC<WorkspaceDataGridProp> = ({
  workspaces,
  maxHeight,
  sx,
  dataGridProps,
}) => {
  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });
  const workspaceInfoDialogDispatch = WorkspaceInfoDialogContext.useDispatch();

  const columns: GridColDef[] = [
    { field: "name", headerName: "Name", flex: 1 },
    { field: "template", headerName: "Template", flex: 1 },
    {
      field: "phase",
      headerName: "Phase",
      flex: 1,
      renderCell: (params: GridRenderCellParams<any, string | undefined>) => (
        <StatusChip label={params.row.phase || "Unknown"} />
      ),
    },
    {
      field: "lastStartedAt",
      headerName: "LastStartedAt",
      valueGetter: (value: Timestamp | undefined) =>
        value?.toDate().toLocaleString() || "-",
      flex: 1,
    },
  ];

  return (
    <>
      <div style={{ width: "100%", maxHeight: maxHeight, minHeight: 100 }}>
        <DataGrid
          autoHeight={maxHeight === undefined}
          rows={workspaces.map((v) => ({
            ...v,
            id: v.name,
            template: v.spec?.template,
            phase: v.status?.phase,
            lastStartedAt: v.status?.lastStartedAt,
          }))}
          columns={columns}
          getRowHeight={() => "auto"}
          sx={{
            [`& .${gridClasses.cell}`]: {
              py: 1,
            },
            ...sx,
          }}
          initialState={{
            columns: {
              columnVisibilityModel: {
                lastStartedAt: isUpSM,
              },
            },
          }}
          onRowDoubleClick={(params) => {
            workspaceInfoDialogDispatch(true, { ws: params.row });
          }}
          hideFooter
          {...dataGridProps}
        />
      </div>
    </>
  );
};
