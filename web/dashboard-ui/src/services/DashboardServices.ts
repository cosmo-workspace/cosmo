import { createPromiseClient } from "@bufbuild/connect";
import { createConnectTransport, createGrpcWebTransport } from "@bufbuild/connect-web";
import { useMemo } from "react";
import { AuthService } from "../proto/gen/dashboard/v1alpha1/auth_service_connectweb";
import { TemplateService } from "../proto/gen/dashboard/v1alpha1/template_service_connectweb";
import { UserService } from "../proto/gen/dashboard/v1alpha1/user_service_connectweb";
import { WorkspaceService } from "../proto/gen/dashboard/v1alpha1/workspace_service_connectweb";

const transportX = createConnectTransport({
    baseUrl: import.meta.env.BASE_URL,
});

const transport = createGrpcWebTransport({
    baseUrl: import.meta.env.BASE_URL,
});

export function useAuthService() {
    return useMemo(() => createPromiseClient(AuthService, transport), [AuthService]);
}
export function useTemplateService() {
    return useMemo(() => createPromiseClient(TemplateService, transport), [TemplateService]);
}
export function useUserService() {
    return useMemo(() => createPromiseClient(UserService, transport), [UserService]);
}

export function useWorkspaceService() {
    return useMemo(() => createPromiseClient(WorkspaceService, transport), [WorkspaceService]);
}