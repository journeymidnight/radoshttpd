#ifndef FORCE_DELETE_H
#define FORCE_DELETE_H

/* force delete the file */
int striprados_remove(rados_ioctx_t ioctx, rados_striper_t striper, char *oid);

#endif
