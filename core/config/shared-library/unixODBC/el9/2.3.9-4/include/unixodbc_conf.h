/*
 * Kluge to support multilib installation of both 32- and 64-bit RPMS:
 * we need to arrange that header files that appear in both RPMs are
 * identical.  Hence, this file is architecture-independent and calls
 * in an arch-dependent file that will appear in just one RPM.
 *
 * To avoid breaking arches not explicitly supported by Red Hat, we
 * use this indirection file *only* on known multilib arches.
 *
 * We pay attention to include _only_ the original multilib-unclean
 * header file.  Including any other system-header file could cause
 * unpredictable include-ordering issues (rhbz#1412274, comment #16).
 *
 * Note: this may well fail if user tries to use gcc's -I- option.
 * But that option is deprecated anyway.
 */
#if defined(__x86_64__)
#include "unixodbc_conf-x86_64.h"
#elif defined(__i386__)
#include "unixodbc_conf-i386.h"
#elif defined(__ppc64__) || defined(__powerpc64__)
#include "unixodbc_conf-ppc64.h"
#elif defined(__ppc__) || defined(__powerpc__)
#include "unixodbc_conf-ppc.h"
#elif defined(__s390x__)
#include "unixodbc_conf-s390x.h"
#elif defined(__s390__)
#include "unixodbc_conf-s390.h"
#elif defined(__sparc__) && defined(__arch64__)
#include "unixodbc_conf-sparc64.h"
#elif defined(__sparc__)
#include "unixodbc_conf-sparc.h"
#endif
