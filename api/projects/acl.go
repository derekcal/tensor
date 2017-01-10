package projects

import (
	"net/http"
	"strconv"

	"bitbucket.pearson.com/apseng/tensor/api/metadata"
	"bitbucket.pearson.com/apseng/tensor/db"
	"bitbucket.pearson.com/apseng/tensor/models"
	"bitbucket.pearson.com/apseng/tensor/roles"
	"bitbucket.pearson.com/apseng/tensor/util"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

// AccessList returns the list of teams and users that is able to access
// current project object in the gin context
func AccessList(c *gin.Context) {
	project := c.MustGet(CTXProject).(models.Project)

	var organization models.Organization
	err := db.Organizations().FindId(project.OrganizationID).One(&organization)
	if err != nil {
		log.Errorln("Error while retriving Organization:", err)
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:     http.StatusInternalServerError,
			Messages: []string{"Error while getting Access List"},
		})
		return
	}

	var allaccess map[bson.ObjectId]*models.AccessType

	// indirect access from organization
	for _, v := range organization.Roles {
		if v.Type == "user" {
			// if an organization admin
			switch v.Role {
			case roles.ORGANIZATION_ADMIN:
				{
					access := gin.H{
						"descendant_roles": []string{
							"admin",
							"execute",
							"read",
						},
						"role": gin.H{
							"resource_name": organization.Name,
							"description":   "Can manage all aspects of the organization",
							"related": gin.H{
								"organization": "/v1/organizations/" + organization.ID.Hex() + "/",
							},
							"resource_type": "organization",
							"name":          roles.ORGANIZATION_ADMIN,
						},
					}

					allaccess[v.UserID].IndirectAccess = append(allaccess[v.UserID].IndirectAccess, access)
				}
			// if an organization auditor or member
			case roles.ORGANIZATION_MEMBER:
				{
					access := gin.H{
						"descendant_roles": []string{
							"execute",
							"read",
						},
						"role": gin.H{
							"resource_name": organization.Name,
							"description":   "Can manage all aspects of the organization",
							"related": gin.H{
								"organization": "/v1/organizations/" + organization.ID.Hex() + "/",
							},
							"resource_type": "organization",
							"name":          roles.ORGANIZATION_MEMBER,
						},
					}

					allaccess[v.UserID].IndirectAccess = append(allaccess[v.UserID].IndirectAccess, access)
				}
			// if an organization auditor
			case roles.ORGANIZATION_AUDITOR:
				{
					access := gin.H{
						"descendant_roles": []string{
							"read",
						},
						"role": gin.H{
							"resource_name": organization.Name,
							"description":   "Can manage all aspects of the organization",
							"related": gin.H{
								"organization": "/v1/organizations/" + organization.ID.Hex() + "/",
							},
							"resource_type": "organization",
							"name":          roles.ORGANIZATION_AUDITOR,
						},
					}
					allaccess[v.UserID].IndirectAccess = append(allaccess[v.UserID].IndirectAccess, access)
				}
			}
		}
	}

	// direct access

	for _, v := range project.Roles {
		if v.Type == "user" {
			// if an job template admin
			switch v.Role {
			case roles.JOB_TEMPLATE_ADMIN:
				{
					access := gin.H{
						"descendant_roles": []string{
							"admin",
							"execute",
							"read",
						},
						"role": gin.H{
							"resource_name": project.Name,
							"description":   "May run the job template",
							"related": gin.H{
								"job_template": "/v1/job_templates/" + project.ID.Hex() + "/",
							},
							"resource_type": "job_template",
							"name":          roles.JOB_TEMPLATE_ADMIN,
						},
					}

					allaccess[v.UserID].DirectAccess = append(allaccess[v.UserID].DirectAccess, access)
				}
			// if an job template execute
			case roles.JOB_TEMPLATE_EXECUTE:
				{
					access := gin.H{
						"descendant_roles": []string{
							"execute",
							"read",
						},
						"role": gin.H{
							"resource_name": project.Name,
							"description":   "Can manage all aspects of the job template",
							"related": gin.H{
								"job_template": "/v1/job_templates/" + project.ID.Hex() + "/",
							},
							"resource_type": "job_template",
							"name":          roles.JOB_TEMPLATE_EXECUTE,
						},
					}
					allaccess[v.UserID].DirectAccess = append(allaccess[v.UserID].DirectAccess, access)
				}
			}
		}

	}

	var usrs []models.AccessUser

	for k, v := range allaccess {
		var user models.AccessUser
		err := db.Users().FindId(k).One(&user)
		if err != nil {
			log.Errorln("Error while retriving user data:", err)
			c.JSON(http.StatusInternalServerError, models.Error{
				Code:     http.StatusInternalServerError,
				Messages: []string{"Error while getting Access List"},
			})
			return
		}

		metadata.AccessUserMetadata(&user)
		user.Summary = v
		usrs = append(usrs, user)
	}

	count := len(usrs)
	pgi := util.NewPagination(c, count)
	//if page is incorrect return 404
	if pgi.HasPage() {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Invalid page " + strconv.Itoa(pgi.Page()) + ": That page contains no results."})
		return
	}
	// send response with JSON rendered data
	c.JSON(http.StatusOK, models.Response{
		Count:    count,
		Next:     pgi.NextPage(),
		Previous: pgi.PreviousPage(),
		Results:  usrs[pgi.Skip():pgi.End()],
	})

}