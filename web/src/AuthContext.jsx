import React from "react";

const defaultState = {
  tenants: {},
  loading: false,
  login: () => null,
};

const AuthContext = React.createContext(defaultState);

export const AuthProvider = ({ children }) => {
  const state = React.useState(defaultState);

  useEffect(() => {}, []);

  return <AuthContext.Provider value={state}>{children}</AuthContext.Provider>;
};

export default AuthContext;
