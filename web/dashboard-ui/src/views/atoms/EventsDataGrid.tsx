import { Info, Warning } from "@mui/icons-material";
import { Chip, SxProps } from "@mui/material";
import {
  DataGrid,
  DataGridProps,
  GridColDef,
  gridClasses,
} from "@mui/x-data-grid";
import React from "react";
import { Event } from "../../proto/gen/dashboard/v1alpha1/event_pb";
import { EventDetailDialogContext } from "../organisms/EventDetailDialog";

export type EventsDataGridProp = {
  events: Event[];
  maxHeight?: number;
  sx?: SxProps;
  dataGridProps?: Omit<DataGridProps, "columns" | "sx">;
  clock: Date;
};

export const EventsDataGrid: React.FC<EventsDataGridProp> = ({
  events,
  maxHeight,
  sx,
  dataGridProps,
  clock,
}) => {
  const eventDetailDialogDispatch = EventDetailDialogContext.useDispatch();

  const columns: GridColDef[] = [
    { field: "type", headerName: "Type", width: 50 },
    {
      field: "eventTime",
      headerName: "LastSeen",
      valueGetter: (value, row: Event) =>
        (row.series?.lastObservedTime || row.eventTime)?.toDate(),
      valueFormatter: (value: Date) =>
        formatTime(clock.getTime() - value.getTime()),
      flex: 0.2,
      minWidth: 80,
    },
    {
      field: "reason",
      headerName: "Reason",
      flex: 0.25,
      renderCell: (params) => (
        <Chip
          icon={
            params.api.getRow(params.id).type == "Normal" ? (
              <Info color="success" />
            ) : (
              <Warning color="warning" />
            )
          }
          label={params.value}
        />
      ),
    },
    { field: "regardingWorkspace", headerName: "Workspace" },
    {
      field: "regarding",
      headerName: "Object",
      valueGetter: (value, row) =>
        row.regarding?.kind + "/" + row.regarding?.name,
      flex: 0.3,
    },
    { field: "reportingController", headerName: "Reporter" },
    {
      field: "series",
      headerName: "Count",
      valueGetter: (value, row) => row.series?.count || 1,
    },
    { field: "note", headerName: "Message", flex: 1 },
  ];

  return (
    <>
      <div style={{ width: "100%", maxHeight: maxHeight, minHeight: 100 }}>
        <DataGrid
          autoHeight={maxHeight === undefined}
          rows={events}
          columns={columns}
          getRowHeight={() => "auto"}
          sx={{
            [`& .${gridClasses.cell}`]: {
              py: 1,
            },
            ...sx,
          }}
          onRowDoubleClick={(params) => {
            eventDetailDialogDispatch(true, { event: params.row });
          }}
          hideFooter
          {...dataGridProps}
        />
      </div>
    </>
  );
};

export function formatTime(milisec: number): string {
  const sec = Math.floor(milisec / 1000) % 60;
  const min = Math.floor(milisec / 1000 / 60) % 60;
  const hours = Math.floor(milisec / 1000 / 60 / 60) % 24;
  if (hours > 0) return `${hours}h${min}m${sec}s`;
  if (min > 0) return `${min}m${sec}s`;
  return `${sec}s`;
}
