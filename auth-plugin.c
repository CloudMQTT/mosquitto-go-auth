#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <errno.h>

#include <mosquitto.h>
#include <mosquitto_plugin.h>
#if MOSQ_AUTH_PLUGIN_VERSION == 3
#include <mosquitto_broker.h>
#endif
#include "go-auth.h"

int mosquitto_auth_plugin_version(void) {
  return MOSQ_AUTH_PLUGIN_VERSION;
}
#if MOSQ_AUTH_PLUGIN_VERSION == 3
int mosquitto_auth_plugin_init(void **user_data, struct mosquitto_opt *auth_opts, int auth_opt_count) {
#else
int mosquitto_auth_plugin_init(void **user_data, struct mosquitto_auth_opt *auth_opts, int auth_opt_count) {
#endif
  /*
    Pass auth_opts hash as keys and values char* arrays to Go in order to initialize them there.
  */

  GoInt32 opts_count = auth_opt_count;
  
  GoString keys[auth_opt_count];
  GoString values[auth_opt_count];
  int i;

#if MOSQ_AUTH_PLUGIN_VERSION == 3
  struct mosquitto_opt *o;
#else
	struct mosquitto_auth_opt *o;
#endif
  for (i = 0, o = auth_opts; i < auth_opt_count; i++, o++) {
    GoString opt_key = {o->key, strlen(o->key)};
    GoString opt_value = {o->value, strlen(o->value)};
    keys[i] = opt_key;
    values[i] = opt_value;
  }

  GoSlice keysSlice = {keys, auth_opt_count, auth_opt_count};
  GoSlice valuesSlice = {values, auth_opt_count, auth_opt_count};

  AuthPluginInit(keysSlice, valuesSlice, opts_count);
  return MOSQ_ERR_SUCCESS;
}

#if MOSQ_AUTH_PLUGIN_VERSION == 3
int mosquitto_auth_plugin_cleanup(void *user_data, struct mosquitto_opt *auth_opts, int auth_opt_count) {
#else
int mosquitto_auth_plugin_cleanup(void *user_data, struct mosquitto_auth_opt *auth_opts, int auth_opt_count) {
#endif
  AuthPluginCleanup();
  return MOSQ_ERR_SUCCESS;
}

#if MOSQ_AUTH_PLUGIN_VERSION == 3
int mosquitto_auth_security_init(void *user_data, struct mosquitto_opt *auth_opts, int auth_opt_count, bool reload) {
#else
int mosquitto_auth_security_init(void *user_data, struct mosquitto_auth_opt *auth_opts, int auth_opt_count, bool reload) {
#endif
  return MOSQ_ERR_SUCCESS;
}

#if MOSQ_AUTH_PLUGIN_VERSION == 3
int mosquitto_auth_security_cleanup(void *user_data, struct mosquitto_opt *auth_opts, int auth_opt_count, bool reload) {
#else
int mosquitto_auth_security_cleanup(void *user_data, struct mosquitto_auth_opt *auth_opts, int auth_opt_count, bool reload) {
#endif
  return MOSQ_ERR_SUCCESS;
}

#if MOSQ_AUTH_PLUGIN_VERSION == 3
int mosquitto_auth_unpwd_check(void *user_data, const struct mosquitto *client, const char *username, const char *password) {
#else
int mosquitto_auth_unpwd_check(void *user_data, const char *username, const char *password) {
#endif
  
  if (username == NULL || password == NULL) {
    printf("error: received null username or password for unpwd check\n");
    fflush(stdout);
    return MOSQ_ERR_AUTH;
  }

  GoString go_username = {username, strlen(username)};
  GoString go_password = {password, strlen(password)};

  if(AuthUnpwdCheck(go_username, go_password)){
    return MOSQ_ERR_SUCCESS;
  }

  return MOSQ_ERR_AUTH;
}
#if MOSQ_AUTH_PLUGIN_VERSION == 3
int mosquitto_auth_acl_check(void *user_data, int access, const struct mosquitto *client, const struct mosquitto_acl_msg *msg) {
#else
int mosquitto_auth_acl_check(void *user_data, const char *clientid, const char *username, const char *topic, int access) {
#endif
	//Try to print everything that gets here so that we can debug issues.
	printf("\n\n");
	printf("auth-plugin.c: starting acl check at auth-plugin.c\n");
#if MOSQ_AUTH_PLUGIN_VERSION == 3
	const char* clientid = mosquitto_client_id(client);
	const char* username = mosquitto_client_username(client);
	const char* topic = msg->topic;
#endif
  if(clientid == NULL) {
    printf("error: clientid is null\n");
  } else {
    printf("clientid: %s\n", clientid);
  }

  if(username == NULL) {
    printf("error: username is null\n");
  } else {
    printf("username: %s\n", username);
  }

  if(topic == NULL) {
    printf("error: topic is null\n");
  } else {
    printf("topic: %s\n", topic);
  }

  if(access < 1) {
    printf("error: access is 0 or negative\n");
  } else {
    printf("access: %d\n", access);
  }

  printf("\n\n");

  fflush(stdout);

  if (clientid == NULL || username == NULL || topic == NULL || access < 1) {
    printf("error: received null username, clientid or topic, or access is equal or less than 0 for acl check\n");
    fflush(stdout);
    return MOSQ_ERR_ACL_DENIED;
  }
  
  GoString go_clientid = {clientid, strlen(clientid)};
  GoString go_username = {username, strlen(username)};
  GoString go_topic = {topic, strlen(topic)};
  GoInt32 go_access = access;

  if(AuthAclCheck(go_clientid, go_username, go_topic, go_access)){
    return MOSQ_ERR_SUCCESS;
  }

  return MOSQ_ERR_ACL_DENIED;
}
#if MOSQ_AUTH_PLUGIN_VERSION == 3
int mosquitto_auth_psk_key_get(void *user_data, const struct mosquitto *client, const char *hint, const char *identity, char *key, int max_key_len) {
#else
int mosquitto_auth_psk_key_get(void *user_data, const char *hint, const char *identity, char *key, int max_key_len) {
#endif
  return MOSQ_ERR_AUTH;
}
