import { Timestamp } from "@bufbuild/protobuf";
import { useState } from "react";
import { ModuleContext } from "../../components/ContextProvider";
import { useHandleError, useLogin } from "../../components/LoginProvider";
import { Event } from "../../proto/gen/dashboard/v1alpha1/event_pb";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { useUserService } from "../../services/DashboardServices";
import { setUserStateFuncFilteredByLoginUserRole } from "./UserModule";
/**
 * hooks
 */
const useEvent = () => {
  console.log("useEvent");
  const { handleError } = useHandleError();

  const [events, setEvents] = useState<Event[]>([]);
  const userService = useUserService();

  const { loginUser, updateClock } = useLogin();
  const [user, setUser] = useState<User>(loginUser || new User());
  const [users, setUsers] = useState<User[]>([loginUser || new User()]);

  const getUsers = async () => {
    console.log("useEvent:getUsers");
    try {
      const result = await userService.getUsers({});
      setUsers(
        setUserStateFuncFilteredByLoginUserRole(result.items, loginUser),
      );
    } catch (error) {
      handleError(error);
    }
  };

  const getEvents = async () => {
    console.log("useEvent:getEvents");
    try {
      const result = await userService.getEvents({ userName: user.name });
      setEvents(result.items);
      updateClock();
    } catch (error) {
      handleError(error);
    }
  };

  return ({
    user,
    setUser,
    users,
    setUsers,
    events,
    setEvents,
    getEvents,
    getUsers,
  });
};

export function getTime(timestamp?: Timestamp): number {
  if (!timestamp) {
    return 0;
  }
  return timestamp.toDate().getTime();
}

export function latestTime(event?: Event): number {
  if (!event) {
    return 0;
  }
  const t1 = getTime(event.series?.lastObservedTime);
  const t2 = getTime(event.eventTime);
  return t1 > t2 ? t1 : t2;
}

export const EventContext = ModuleContext(useEvent);
export const useEventModule = EventContext.useContext;
