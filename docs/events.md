![Lostrómos logo](images/logo.png)

# Understanding Lostrómos Events

This document is meant to describe the events that will occur when you use
 Lostrómos with different settings.

## Updates with a Filter

When performing updates and using a non empty filter via the `crd.filter`
 option, we have defined the behavior:

| Old Resource | New Resource | Action Taken |
| ------------ | ------------ | ------------ |
| Filter Annotation Exists | Filter Annotation Exists | ResourceUpdated |
| Filter Annotation Doesn't Exist | Filter Annotation Exists | ResourceAdded |
| Filter Annotation Exists | Filter Annotation Doesn't Exist | ResourceDeleted |
| Filter Annotation Doesn't Exist | Filter Annotation Doesn't Exist | No-Op |

In the case that filtering isn't used, `ResourceUpdated` is called.