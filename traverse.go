package liqu

import (
	"fmt"
	"strings"
)

/*
*
SELECT json_agg(q)
FROM

	(SELECT count(*) OVER() AS TotalRows,
	        row_to_json(project) AS "Project",
	        ProjectTags."ProjectTags",
	        ProjectAttributes."ProjectAttributes"
	 FROM
	   (SELECT project.id AS "ID",
	           project.title AS "Title"
	    FROM project) AS "project"
	 LEFT JOIN LATERAL
	   (SELECT COALESCE(jsonb_agg(jsonb_build_object('ProjectID', project_tag.id_project, 'TagID', project_tag.id_tag, 'Tags', Tags."Tags")), '[]') AS "ProjectTags"
	    FROM project_tag
	    LEFT JOIN LATERAL
	      (SELECT COALESCE(jsonb_agg(jsonb_build_object('Tag', tag.tag, 'Color', tag.color, 'ID', tag.id)), '[]') AS "Tags"
	       FROM tag
	       WHERE tag.id=project_tag."id_tag" ) AS Tags ON TRUE
	    WHERE project_tag.id_project=project."ID" ) AS ProjectTags ON TRUE
	 LEFT JOIN LATERAL
	   (SELECT COALESCE(jsonb_agg(jsonb_build_object('ProjectID', project_attribute.id_project, 'AttributeID', project_attribute.id_attribute, 'Attribute', row_to_json(Attribute), 'ProjectAttributeTag', row_to_json(ProjectAttributeTag))), '[]') AS "ProjectAttributes"
	    FROM project_attribute
	    LEFT JOIN LATERAL
	      (SELECT jsonb_build_object('Name', attribute.name, 'ID', attribute.id) AS "Attribute",
	              id
	       FROM attribute
	       WHERE attribute.id=project_attribute."id_attribute" ) AS Attribute ON TRUE
	    LEFT JOIN LATERAL
	      (SELECT jsonb_build_object('ProjectID', project_attribute_tag.id_project, 'AttributeID', project_attribute_tag.id_attribute) AS "ProjectAttributeTag",
	              id_tag
	       FROM project_attribute_tag
	       WHERE attribute."id"=project_attribute_tag.id_attribute
	         AND project."ID"=project_attribute_tag.id_project ) AS ProjectAttributeTag ON TRUE
	    WHERE project_attribute.id_project=project."ID" ) AS ProjectAttributes ON TRUE
	 GROUP BY project.*,
	          ProjectTags."ProjectTags",
	          ProjectAttributes."ProjectAttributes"
	 LIMIT 99999
	 OFFSET 0) q
*/

func (l *Liqu) traverse() error {
	root := NewRootQuery()

	if l.sourceSlice {
		root.SetTotalRows("count(*) OVER() AS TotalRows,")
	}

	if !l.tree.anonymous {
		baseSelect := fmt.Sprintf("to_jsonb( :select: ) AS %s", l.tree.As)
		if l.tree.slice {
			baseSelect = fmt.Sprintf("jsonb_agg( :select: ) AS %s", l.tree.As)
		}
		root.setSelect(baseSelect)
	}

	base := NewBaseQuery()
	base.setFrom(l.tree.registry.tableName)
	base.setSelect(strings.Join(l.selectsWithStructAlias(&l.tree), ","))

	var rootSelects []string

	if l.tree.anonymous {
		rootSelects = l.selectsAsStruct(&l.tree)
		root.setSelect(strings.Join(l.selectsAsStruct(&l.tree), ","))
	} else {
		rootSelects = []string{l.tree.As}

	}

	root.setSelect(strings.Join(rootSelects, ",")).
		setFrom(base.Scrub()).
		setAs(l.tree.As, l.tree.registry.tableName)

	for _, v := range l.tree.branches {
		err := l.traverseBranch(v)
		if err != nil {
			return err
		}
	}

	l.sqlQuery = root.Scrub()

	return nil
}

func (l *Liqu) traverseBranch(branch *branch) error {
	Debug(branch)
	for _, v := range branch.branches {
		err := l.traverseBranch(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Liqu) selectsAsStruct(branch *branch) []string {
	var out []string

	for _, field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`%s."%s"`, branch.Name, field))
	}

	return out
}

func (l *Liqu) selectsAsPair(branch *branch) []string {
	var out []string

	for _, field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`'%s'`, field), fmt.Sprintf(`%s."%s"`, branch.source.Table(), l.registry[branch.Name].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) selectsWithStructAlias(branch *branch) []string {
	var out []string

	for _, field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`%s.%s AS "%s"`, branch.source.Table(), l.registry[branch.Name].fieldDatabase[field], field))
	}

	return out
}
